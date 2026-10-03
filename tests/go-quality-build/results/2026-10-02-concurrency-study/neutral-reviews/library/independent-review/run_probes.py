#!/usr/bin/env python3
import json
from pathlib import Path
import shutil
from run_checks import PACKET, OUT, GO122, copy_candidate, run, hashes

observations = []
mutations = []
for c in ['A', 'B']:
    work = copy_candidate(c, 'contract-' + c)
    shutil.copy2(PACKET / 'held-checks' / 'contract_test.go', work / 'held_contract_test.go')
    shutil.copy2(OUT / 'contract_probe_test.go', work / 'review_contract_test.go')
    for go, suffix in [('go', 'host'), (GO122, 'go122')]:
        observations.append(run(c + '-contract-' + suffix, work,
            [go, 'test', '-mod=readonly', '-race', '-shuffle=on', '-count=3', '-timeout=15s', './...']))

    for name in ['drop-unit-updates', 'split-update', 'global-state', 'wrong-signed-sum']:
        mutant = copy_candidate(c, name + '-' + c)
        source = (mutant / 'totals.go').read_text()
        if name == 'drop-unit-updates':
            source = source.replace('t.count++', 'if delta == 1 {\n\t\tt.mu.Unlock()\n\t\treturn\n\t}\n\tt.count++')
            pattern = '^TestConcurrentAdd'
            intended = 'The final exact-count assertion should fail for coherent {0 0}, with no wait for the expected count.'
        elif name == 'split-update':
            source = source.replace('import "sync"', 'import ("sync"; "runtime")')
            source = source.replace('t.count++\n\tt.sum += delta',
                't.count++\n\tt.mu.Unlock()\n\truntime.Gosched()\n\tt.mu.Lock()\n\tt.sum += delta')
            pattern = '^TestConcurrentAdd'
            intended = 'The in-flight Count != Sum assertion should fail; every field remains separately synchronized, so race rejection is not required.'
        elif name == 'global-state':
            source = source.replace('// Add accepts one update.', 'var reviewGlobal Totals\n\n// Add accepts one update.')
            source = source.replace('func (t *Totals) Add(delta int64) {', 'func (t *Totals) Add(delta int64) {\n\tt = &reviewGlobal')
            source = source.replace('func (t *Totals) Snapshot() Snapshot {', 'func (t *Totals) Snapshot() Snapshot {\n\tt = &reviewGlobal')
            pattern = 'Independent|Isolated'
            intended = 'The first or second instance snapshot assertion should reject state shared between Totals instances.'
        else:
            source = source.replace('t.sum += delta', 'if delta < 0 { delta = -delta }; t.sum += delta')
            pattern = '^TestSequentialTotals$'
            intended = 'The sequential exact {Count:2, Sum:1} assertion should reject treating a negative delta as positive.'
        (mutant / 'totals.go').write_text(source)
        for go, suffix in [('go', 'host'), (GO122, 'go122')]:
            key = c + '-mutation-' + name + '-' + suffix
            record = run(key, mutant, [go, 'test', '-mod=readonly', '-count=1', '-timeout=2s', '-run', pattern, './...'])
            output = record['stdout'] + record['stderr']
            assertion = ('--- FAIL:' in output and 'test timed out' not in output and '[build failed]' not in output)
            mutations.append({'candidate': c, 'mutation': name, 'toolchain': suffix,
                'intended_assertion': intended, 'behavioral_assertion_detected': assertion,
                'deadline_rejection': 'test timed out' in output,
                'build_failure': '[build failed]' in output,
                'raw_record': 'raw/' + key + '.json', 'source': str(mutant / 'totals.go')})

(OUT / 'mutation-sensitivity.json').write_text(json.dumps(mutations, indent=2) + '\n')
(OUT / 'source-hashes-after-probes.json').write_text(json.dumps(hashes(), indent=2) + '\n')
