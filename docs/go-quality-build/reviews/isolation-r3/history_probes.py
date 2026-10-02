#!/usr/bin/env python3
"""Reproduce selected historical/current boundaries without altering evidence."""
import json
import os
from pathlib import Path
import shutil
import sys
import tempfile

sys.path.insert(0, '/private/tmp/go-isolation-r3-promotion')
from verify import ROOT, OUT, RESULTS, call, inventory

records = []
env = dict(os.environ, GOCACHE=str(OUT / 'cache'), GOTOOLCHAIN='local')
def run(args, cwd, label, chosen=env):
    r = call(args, cwd, chosen, 180)
    r['label'] = label
    r['env'] = {'GOCACHE': chosen.get('GOCACHE', '<unset>'), 'GOTOOLCHAIN': chosen.get('GOTOOLCHAIN'), 'HOME': chosen.get('HOME', '<unset>')}
    records.append(r)
    (OUT / 'executed-history-probes.json').write_text(json.dumps(records, indent=2) + '\n')
    print(label, r['exit_code'], flush=True)
    return r

redirect = RESULTS / '2026-10-01-testing-combined-r3/isolation-only/post-review-diagnostic/redirect_test.go'
cause = RESULTS / '2026-10-01-testing-combined-r3/cause-crosscheck/2026-10-01-testing-combined-r3--isolation-only/diagnostic_cause_test.go'
targets = [('2026-10-01-testing-combined', x) for x in ['architecture-only', 'behavior-only', 'isolation-only', 'both']]
targets += [('2026-10-01-testing-combined-r2', x) for x in ['isolation-only', 'both']]
targets += [('2026-10-01-testing-combined-r3', x) for x in ['isolation-only', 'both']]
for name, case in targets:
    with tempfile.TemporaryDirectory(prefix='boundary-history-', dir=OUT) as tmp:
        work = Path(tmp) / 'module'
        source = RESULTS / name / case / 'candidate'
        shutil.copytree(source, work)
        shutil.copy2(redirect, work / 'reviewer_redirect_test.go')
        run(['go', 'test', '-run', '^TestDiagnosticSingleRequestInitialStatus$', '-count=1', '-timeout=30s', './...'], work, name + '/' + case + ' concrete-client boundary')
        if name.endswith(('r2', 'r3')):
            shutil.copy2(cause, work / 'reviewer_cause_test.go')
            run(['go', 'test', '-run', '^TestDiagnosticNonComparableCause$', '-count=1', '-timeout=30s', './...'], work, name + '/' + case + ' valid noncomparable cause no-panic/order/independent-failure')

with tempfile.TemporaryDirectory(prefix='historical-default-', dir=OUT) as tmp:
    work = Path(tmp) / 'module'
    shutil.copytree(RESULTS / '2026-10-01-testing-combined-final/both/candidate', work)
    exe = Path(tmp) / 'command-tests'
    run(['go', 'test', '-c', '-o', str(exe), './cmd/indexer'], work, 'historical behavior2/isolation2 both compile unchanged command tests')
    runtime = dict(env, HOME=str(OUT / 'valid-default-home'))
    runtime.pop('GOCACHE', None)
    run([str(exe), '-test.count=1', '-test.timeout=60s'], work / 'cmd/indexer', 'historical behavior2/isolation2 both unset-GOCACHE valid-HOME setup', runtime)
