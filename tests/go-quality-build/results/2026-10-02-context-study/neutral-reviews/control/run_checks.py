from concurrent.futures import ThreadPoolExecutor
from datetime import datetime, timezone
from hashlib import sha256
import json
from pathlib import Path
import shutil
import subprocess

ROOT = Path('/private/tmp/go-context-outcome-20261002/control')
OUT = ROOT / 'reviews'
FILES = ('README.md', 'go.mod', 'range.go', 'range_test.go')

def hashes(area):
    return {name: sha256((ROOT / area / name).read_bytes()).hexdigest() for name in FILES}

original_hashes = hashes('original')
before = {f'candidate-{n}': hashes(f'candidate-{n}') for n in (1, 2)}
skill_hashes = {str(p.relative_to(ROOT)): sha256(p.read_bytes()).hexdigest()
                for p in sorted((ROOT / 'review-skills').rglob('*')) if p.is_file()}

boundary_probe = '''package rangecheck

import "testing"

func TestIndependentBoundaries(t *testing.T) {
    max := int(^uint(0) >> 1)
    min := -max - 1
    cases := []struct {
        name string
        value, low, high int
        want bool
    }{
        {"singleton included", 3, 3, 3, true},
        {"singleton below", 2, 3, 3, false},
        {"singleton above", 4, 3, 3, false},
        {"negative lower", -6, -6, -2, true},
        {"negative upper", -2, -6, -2, true},
        {"negative interior", -4, -6, -2, true},
        {"negative below", -7, -6, -2, false},
        {"negative above", -1, -6, -2, false},
        {"zero singleton", 0, 0, 0, true},
        {"minimum singleton", min, min, min, true},
        {"maximum singleton", max, max, max, true},
        {"full range minimum", min, min, max, true},
        {"full range maximum", max, min, max, true},
        {"full range zero", 0, min, max, true},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            if got := inRange(c.value, c.low, c.high); got != c.want {
                t.Fatalf("inRange(%d, %d, %d) = %v; want %v", c.value, c.low, c.high, got, c.want)
            }
        })
    }
}
'''

def review(n):
    name = f'candidate-{n}'
    dest = OUT / name
    evidence = dest / 'evidence'
    scratch = dest / 'scratch'
    evidence.mkdir(parents=True, exist_ok=True)
    scratch.mkdir(parents=True, exist_ok=True)
    copies = {}
    for tag in ('suite', 'probes', 'original-with-regression'):
        path = scratch / tag
        shutil.copytree(ROOT / name, path)
        copies[tag] = path
    (copies['probes'] / 'contract_test.go').write_bytes((ROOT / 'provided-contract-probes/contract_test.go').read_bytes())
    (copies['probes'] / 'independent_boundaries_test.go').write_text(boundary_probe)
    (copies['original-with-regression'] / 'range.go').write_bytes((ROOT / 'original/range.go').read_bytes())
    cache = scratch / 'go-cache'
    modcache = scratch / 'go-mod-cache'
    temp = scratch / 'go-temp'
    for path in (cache, modcache, temp):
        path.mkdir()
    env_args = ['GOWORK=off', 'GOTOOLCHAIN=local', 'GOPROXY=off', 'GOSUMDB=off',
                f'GOCACHE={cache}', f'GOMODCACHE={modcache}', f'GOTMPDIR={temp}', 'GOFLAGS=']
    checks = [
        ('version', 'suite', ['go', 'version'], 0),
        ('suite', 'suite', ['go', 'test', '-count=1', '-mod=readonly', './...'], 0),
        ('vet', 'suite', ['go', 'vet', '-mod=readonly', './...'], 0),
        ('format', 'suite', ['gofmt', '-d', 'range.go', 'range_test.go'], 0),
        ('module-graph', 'suite', ['go', 'list', '-mod=readonly', '-m', 'all'], 0),
        ('language-1.22', 'suite', ['go', 'test', '-count=1', '-mod=readonly', '-gcflags=example.com/rangecheck=-lang=go1.22', './...'], 0),
        ('contract-and-boundaries', 'probes', ['go', 'test', '-count=1', '-mod=readonly', '-v', './...'], 0),
        ('regression-rejects-original', 'original-with-regression', ['go', 'test', '-count=1', '-mod=readonly', '-run=TestUpper', '-v', './...'], 1),
    ]
    results = []
    for label, tag, args, expected in checks:
        command = ['rtk', 'proxy', 'env', *env_args, *args]
        completed = subprocess.run(command, cwd=copies[tag], text=True, capture_output=True)
        log = evidence / f'{label}.txt'
        log.write_text(f'cwd: {copies[tag]}\ncommand argv: {json.dumps(command)}\nexit code: {completed.returncode}\nexpected exit code: {expected}\n\nstdout:\n{completed.stdout}\nstderr:\n{completed.stderr}')
        results.append({'check': label, 'cwd': str(copies[tag]), 'command': command,
                        'exit_code': completed.returncode, 'expected_exit_code': expected,
                        'expected_outcome': completed.returncode == expected,
                        'stdout': completed.stdout, 'stderr': completed.stderr,
                        'log': str(log)})
    after = hashes(name)
    result = {'candidate': name, 'executed_at_utc': datetime.now(timezone.utc).isoformat(),
              'source_hashes_before': before[name], 'source_hashes_after': after,
              'source_unchanged': before[name] == after, 'checks': results,
              'all_expected_outcomes': all(r['expected_outcome'] for r in results),
              'limitations': ['Runtime execution used Go 1.26.5 darwin/arm64, not an actual Go 1.22 toolchain.',
                              'The -lang=go1.22 check verifies the package language level using the host compiler.',
                              'Original provided-checks JSON files were not used; all reported commands were executed independently.',
                              'Only packet source is in scope; no external callers, CI, release, or deployment environment was supplied.']}
    (dest / 'independent-checks.json').write_text(json.dumps(result, indent=2) + '\n')
    return {'candidate': name, 'all_expected_outcomes': result['all_expected_outcomes'],
            'source_unchanged': result['source_unchanged'],
            'checks': [{k: r[k] for k in ('check', 'exit_code', 'expected_exit_code', 'stdout', 'stderr')} for r in results]}

with ThreadPoolExecutor(max_workers=2) as pool:
    results = list(pool.map(review, (1, 2)))

snapshot = {'created_at_utc': datetime.now(timezone.utc).isoformat(),
            'original_hashes': original_hashes, 'candidate_hashes': before,
            'review_contract_hashes': skill_hashes,
            'original_unchanged': original_hashes == hashes('original'),
            'review_contracts_unchanged': all(sha256((ROOT / path).read_bytes()).hexdigest() == expected for path, expected in skill_hashes.items())}
(OUT / 'snapshot.json').write_text(json.dumps(snapshot, indent=2) + '\n')
print(json.dumps({'results': results, 'original_unchanged': snapshot['original_unchanged'],
                  'review_contracts_unchanged': snapshot['review_contracts_unchanged']}, indent=2))
