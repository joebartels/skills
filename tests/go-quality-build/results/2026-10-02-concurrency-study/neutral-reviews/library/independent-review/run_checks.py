#!/usr/bin/env python3
"""Independent, source-preserving review checks. Every command runs through RTK."""
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import time

PACKET = Path('/private/tmp/go-neutral-library-study')
OUT = PACKET / 'independent-review'
GO122 = '/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go'
OUT.mkdir(exist_ok=True)
(OUT / 'raw').mkdir(exist_ok=True)
(OUT / 'work').mkdir(exist_ok=True)

def hashes():
    result = {}
    for candidate in ['original', 'candidates/A', 'candidates/B']:
        result[candidate] = {p.name: hashlib.sha256(p.read_bytes()).hexdigest()
                             for p in sorted((PACKET / candidate).iterdir()) if p.is_file()}
    return result

def run(key, cwd, cmd, env_extra=None, timeout=60):
    env_updates = {'GOTOOLCHAIN': 'local', 'GOWORK': 'off', 'GOPROXY': 'off',
                   'GOSUMDB': 'off', 'GOMODCACHE': str(OUT / 'modcache'),
                   'GOCACHE': str(OUT / 'gocache')}
    env_updates.update(env_extra or {})
    argv = ['rtk', 'proxy', *cmd]
    started = time.monotonic()
    try:
        proc = subprocess.run(argv, cwd=str(cwd), env={**os.environ, **env_updates},
                              text=True, capture_output=True, timeout=timeout)
        status, stdout, stderr = proc.returncode, proc.stdout, proc.stderr
        timed_out = False
    except subprocess.TimeoutExpired as exc:
        status = None
        stdout = exc.stdout.decode() if isinstance(exc.stdout, bytes) else (exc.stdout or '')
        stderr = exc.stderr.decode() if isinstance(exc.stderr, bytes) else (exc.stderr or '')
        timed_out = True
    record = {'id': key, 'argv': argv, 'cwd': str(cwd), 'environment_overrides': env_updates,
              'status': status, 'timed_out': timed_out, 'stdout': stdout, 'stderr': stderr,
              'elapsed_seconds': round(time.monotonic() - started, 3)}
    (OUT / 'raw' / (key + '.json')).write_text(json.dumps(record, indent=2) + '\n')
    (OUT / 'raw' / (key + '.txt')).write_text('COMMAND: ' + ' '.join(argv) + '\nCWD: ' + str(cwd) +
        '\nENV: ' + json.dumps(env_updates, sort_keys=True) + '\nSTDOUT:\n' + stdout + '\nSTDERR:\n' + stderr +
        '\nSTATUS: ' + str(status) + '\nTIMED_OUT: ' + str(timed_out) + '\n')
    print(json.dumps({'id': key, 'status': status, 'timed_out': timed_out,
                      'elapsed_seconds': record['elapsed_seconds'],
                      'result': (stdout + stderr)[-800:]}), flush=True)
    return record

def copy_candidate(candidate, name):
    dest = OUT / 'work' / name
    if dest.exists():
        shutil.rmtree(dest)
    shutil.copytree(PACKET / 'candidates' / candidate, dest)
    return dest

if __name__ == '__main__':
    (OUT / 'source-hashes-before.json').write_text(json.dumps(hashes(), indent=2) + '\n')
    for c in ['A', 'B']:
        work = copy_candidate(c, 'baseline-' + c)
        for go, suffix in [('go', 'host'), (GO122, 'go122')]:
            run(c + '-version-' + suffix, work, [go, 'version'])
            run(c + '-build-' + suffix, work, [go, 'build', '-mod=readonly', '-o', '/dev/null', './...'])
            run(c + '-test-' + suffix, work, [go, 'test', '-mod=readonly', '-count=1', '-timeout=10s', './...'])
            run(c + '-vet-' + suffix, work, [go, 'vet', '-mod=readonly', './...'])
            run(c + '-race-' + suffix, work, [go, 'test', '-mod=readonly', '-race', '-shuffle=on',
                                            '-count=5', '-timeout=15s', './...'])
        run(c + '-gofmt', work, ['gofmt', '-d', 'totals.go', 'totals_test.go'])
        run(c + '-modules', work, [GO122, 'list', '-mod=readonly', '-m', 'all'])
    (OUT / 'source-hashes-after-baseline.json').write_text(json.dumps(hashes(), indent=2) + '\n')
