from pathlib import Path
import json
import os
import subprocess
import time

PACKET = Path('/private/tmp/go-context-outcome-20261002/library')
PATH = PACKET / 'reviews/checks.json'
data = json.loads(PATH.read_text())
scratch = Path(data['scratch'])
minimum = '/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go'
env_changes = dict(GOWORK='off', GOTOOLCHAIN='local', GOPROXY='off', GOCACHE=str(scratch/'go122-gocache'), GOMODCACHE=str(scratch/'go122-gomodcache'))

def run(label, args, cwd):
    argv = ['rtk', 'proxy', minimum] + args
    start = time.monotonic()
    p = subprocess.run(argv, cwd=cwd, env=dict(os.environ, **env_changes), text=True, capture_output=True, timeout=90)
    row = dict(label=label, argv=argv, cwd=str(cwd), environment_overrides=env_changes, status=p.returncode, stdout=p.stdout, stderr=p.stderr, elapsed_seconds=round(time.monotonic()-start, 3))
    data['checks'].append(row)
    PATH.write_text(json.dumps(data, indent=2)+'\n')
    print(label, 'status='+str(p.returncode), p.stdout+p.stderr, flush=True)

run('minimum-toolchain-version', ['version'], PACKET)
for n in (1, 2):
    run(f'candidate-{n}/go122-build', ['build', '-mod=readonly', './...'], scratch/f'candidate-{n}-author')
    run(f'candidate-{n}/go122-author', ['test', '-count=1', '-timeout=30s', './...'], scratch/f'candidate-{n}-author')
    run(f'candidate-{n}/go122-provided-contract', ['test', '-count=1', '-timeout=30s', './...'], scratch/f'candidate-{n}-contract')
