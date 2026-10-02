import hashlib
import json
import os
import pathlib
import shutil
import subprocess
import time

BASE = pathlib.Path('/private/tmp/go-independent-review-4qorzxhe')
OUT = BASE / 'output'
OUT.mkdir(exist_ok=True)
ENV = dict(os.environ, GOCACHE='/private/tmp/go-quality-testing-cache', GOTOOLCHAIN='local', GOPROXY='off')
checks = []

def check(name, argv, cwd, timeout=60):
    started = time.monotonic()
    try:
        p = subprocess.run(argv, cwd=cwd, env=ENV, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=timeout)
        result = dict(name=name, argv=argv, cwd=str(cwd), environment_overrides={k: ENV[k] for k in ('GOCACHE','GOTOOLCHAIN','GOPROXY')}, stdout=p.stdout, stderr=p.stderr, exit_code=p.returncode, elapsed_seconds=time.monotonic()-started, timeout_seconds=timeout)
    except subprocess.TimeoutExpired as exc:
        result = dict(name=name, argv=argv, cwd=str(cwd), environment_overrides={k: ENV[k] for k in ('GOCACHE','GOTOOLCHAIN','GOPROXY')}, stdout=exc.stdout.decode() if isinstance(exc.stdout, bytes) else exc.stdout or '', stderr=exc.stderr.decode() if isinstance(exc.stderr, bytes) else exc.stderr or '', exit_code=None, elapsed_seconds=time.monotonic()-started, timeout_seconds=timeout, timed_out=True)
    checks.append(result)
    (OUT / 'checks.json').write_text(json.dumps(checks, indent=2)+'\n')
    print(json.dumps({k: result[k] for k in ('name','exit_code','elapsed_seconds','stdout','stderr')}), flush=True)
    return result

def go(name, cwd, *args, timeout=60):
    return check(name, ['rtk','proxy','go',*args], cwd, timeout)

if __name__ == '__main__':
    go('toolchain', BASE / 'candidate', 'version')
    go('runtime', BASE / 'candidate', 'env', 'GOOS','GOARCH','GOVERSION','CGO_ENABLED','GOROOT')
    go('original-baseline', BASE / 'original', 'test','-count=1','-timeout=20s','./...')
    go('candidate-normal', BASE / 'candidate', 'test','-count=1','-timeout=20s','./...')
    go('candidate-race-repeat-shuffle', BASE / 'candidate', 'test','-race','-count=10','-shuffle=20261001','-timeout=40s','./...')
    go('candidate-alpha-alone', BASE / 'candidate', 'test','-race','-count=10','-run=^TestIndependentInvocations$/^alpha$','-timeout=20s','./...')
    go('candidate-beta-alone', BASE / 'candidate', 'test','-race','-count=10','-run=^TestIndependentInvocations$/^beta$','-timeout=20s','./...')
    go('candidate-vet-version', BASE / 'candidate', 'vet','-stdversion','./...')
    check('production-contract-config-unchanged', ['rtk','proxy','python3','-c', 'import hashlib,pathlib; b=pathlib.Path("/private/tmp/go-independent-review-4qorzxhe"); [(print(f,hashlib.sha256((b/"original"/f).read_bytes()).hexdigest()), None if (b/"original"/f).read_bytes()==(b/"candidate"/f).read_bytes() else (_ for _ in ()).throw(AssertionError(f))) for f in ("poll.go","go.mod","README.md")]'], BASE)
