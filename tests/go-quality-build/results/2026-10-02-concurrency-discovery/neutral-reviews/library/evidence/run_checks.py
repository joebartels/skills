from pathlib import Path
import hashlib
import json
import os
import subprocess
import time

ROOT = Path('/private/tmp/go-independent-concurrency-library')
EVIDENCE = ROOT / 'evidence'
CHECKS = ROOT / 'disposable-checks'
MIN_GO = '/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go'
env = dict(os.environ)
overrides = {
    'GOCACHE': str(EVIDENCE / 'go-cache'),
    'GOMODCACHE': str(EVIDENCE / 'go-module-cache'),
    'GOTOOLCHAIN': 'local',
    'GOWORK': 'off',
    'GOPROXY': 'off',
    'GOSUMDB': 'off',
    'GOFLAGS': '-mod=readonly',
}
env.update(overrides)
results = []

def run(check_id, subdir, args, expected=0):
    cwd = CHECKS / subdir
    argv = ['rtk', 'proxy', *args]
    began = time.monotonic()
    result = subprocess.run(argv, cwd=cwd, env=env, capture_output=True, text=True, timeout=120)
    record = {
        'id': check_id,
        'argv': argv,
        'cwd': str(cwd),
        'environment_overrides': overrides,
        'status': result.returncode,
        'expected_status': expected,
        'stdout': result.stdout,
        'stderr': result.stderr,
        'elapsed_seconds': round(time.monotonic() - began, 3),
    }
    results.append(record)
    (EVIDENCE / 'executed-checks.json').write_text(json.dumps(results, indent=2) + '\n')
    (EVIDENCE / (check_id + '.log')).write_text(result.stdout + result.stderr)
    print(f'{check_id}: status={result.returncode}; expected={expected}; {record["elapsed_seconds"]}s', flush=True)
    if check_id in ('go-version', 'minimum-go-version', 'module-graph', 'nonstandard-packages'):
        print(result.stdout.rstrip(), flush=True)
    if result.returncode != expected:
        print((result.stdout + result.stderr)[:3000], flush=True)

run('go-version', 'candidate', ['go', 'version'])
run('minimum-go-version', 'candidate', [MIN_GO, 'version'])
run('module-graph', 'candidate', ['go', 'list', '-m', 'all'])
run('nonstandard-packages', 'candidate', ['go', 'list', '-deps', '-f', '{{if not .Standard}}{{.ImportPath}}{{end}}', './...'])
run('build', 'candidate', ['go', 'build', '-o', '/dev/null', './...'])
run('ordinary-tests', 'candidate', ['go', 'test', '-count=1', '-timeout=30s', './...'])
run('vet', 'candidate', ['go', 'vet', './...'])
run('format-diff', 'candidate', ['gofmt', '-d', '.'])
run('race-shuffle-tests', 'candidate', ['go', 'test', '-race', '-shuffle=on', '-count=3', '-timeout=60s', './...'])
run('minimum-version-tests', 'candidate', [MIN_GO, 'test', '-count=1', '-timeout=30s', './...'])
run('independent-contract-race-tests', 'candidate-contract', ['go', 'test', '-race', '-shuffle=on', '-count=3', '-timeout=60s', './...'])
run('minimum-version-contract-tests', 'candidate-contract', [MIN_GO, 'test', '-count=1', '-timeout=30s', './...'])
run('original-negative-control', 'original-with-candidate-tests', ['go', 'test', '-race', '-run=TestConcurrentAddAndSnapshot', '-count=1', '-timeout=30s', './...'], expected=1)

hashes = {str(path.relative_to(ROOT)): hashlib.sha256(path.read_bytes()).hexdigest()
          for group in ('original', 'candidate')
          for path in sorted((ROOT / group).rglob('*')) if path.is_file()}
before = json.loads((EVIDENCE / 'source-hashes-before.json').read_text())
assert hashes == before, 'reviewed source changed'
(EVIDENCE / 'source-hashes-after.json').write_text(json.dumps(hashes, indent=2) + '\n')
print('original and candidate hashes unchanged', flush=True)
