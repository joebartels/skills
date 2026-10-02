import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import time

ROOT = Path('/private/tmp/go-context-outcome-20261002/transfer')
OUT = ROOT / 'reviews'
SCRATCH = OUT / 'diagnostics'
MIN = '/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go'
HOST = shutil.which('go')
OUT.mkdir(exist_ok=True)
SCRATCH.mkdir(exist_ok=True)

def digest_tree(path):
    return {str(p.relative_to(path)): hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(path.rglob('*')) if p.is_file()}

snapshots = {f'candidate-{n}': digest_tree(ROOT / f'candidate-{n}') for n in (1, 2)}
snapshots['original'] = digest_tree(ROOT / 'original')
snapshots['provided-contract-probes'] = digest_tree(ROOT / 'provided-contract-probes')
(OUT / 'source-snapshots.json').write_text(json.dumps(snapshots, indent=2) + '\n')

def run(label, argv, cwd, toolchain, timeout=60):
    env = dict(os.environ)
    env.update(GOTOOLCHAIN='local', GOWORK='off', GOCACHE=str(SCRATCH / 'cache'), GOQUALITY_GO=toolchain)
    env['PATH'] = str(Path(toolchain).parent) + os.pathsep + os.environ['PATH']
    started = time.monotonic()
    command = ['rtk', 'proxy'] + argv
    try:
        result = subprocess.run(command, cwd=cwd, env=env, text=True, capture_output=True, timeout=timeout)
        status, stdout, stderr = result.returncode, result.stdout, result.stderr
    except subprocess.TimeoutExpired as exc:
        status = 'controller_timeout'
        stdout = exc.stdout.decode() if isinstance(exc.stdout, bytes) else exc.stdout or ''
        stderr = exc.stderr.decode() if isinstance(exc.stderr, bytes) else exc.stderr or ''
    record = dict(label=label, argv=command, cwd=str(cwd), environment_overrides={k: env[k] for k in ('GOTOOLCHAIN', 'GOWORK', 'GOCACHE', 'GOQUALITY_GO', 'PATH')}, status=status, stdout=stdout, stderr=stderr, elapsed_seconds=round(time.monotonic()-started, 3))
    (OUT / f'{label}.json').write_text(json.dumps(record, indent=2) + '\n')
    print(json.dumps({k: record[k] for k in ('label', 'status', 'stdout', 'stderr', 'elapsed_seconds')}), flush=True)
    return record

run('host-version', [HOST, 'version'], ROOT, HOST)
run('minimum-version', [MIN, 'version'], ROOT, MIN)
for n in (1, 2):
    target = SCRATCH / f'candidate-{n}'
    if target.exists():
        shutil.rmtree(target)
    shutil.copytree(ROOT / f'candidate-{n}', target)
    for name, compiler in [('host', HOST), ('minimum', MIN)]:
        run(f'candidate-{n}-{name}-author-tests', [compiler, 'test', '-count=1', '-timeout=30s', './...'], target, compiler)
        run(f'candidate-{n}-{name}-build', [compiler, 'build', '-mod=readonly', './...'], target, compiler)
        run(f'candidate-{n}-{name}-vet', [compiler, 'vet', './...'], target, compiler)
    run(f'candidate-{n}-gofmt', ['gofmt', '-d', '.'], target, HOST)
    shutil.copy2(ROOT / 'provided-contract-probes/contract_test.go', target / 'provided_contract_test.go')
    for name, compiler in [('host', HOST), ('minimum', MIN)]:
        run(f'candidate-{n}-{name}-provided-probes', [compiler, 'test', '-run', '^(TestFinalizationAfterCancel|TestFinalizationDeadline|TestUnacceptedCanceledWork|TestFinalizationAfterCancelProcess|TestRequiredReceiptFailureProcess)$', '-count=1', '-timeout=30s', './...'], target, compiler)
    run(f'candidate-{n}-host-race', [HOST, 'test', '-race', '-shuffle=on', '-count=3', '-timeout=60s', './...'], target, HOST, timeout=90)

after = {key: digest_tree(ROOT / key) for key in snapshots}
(OUT / 'source-preservation.json').write_text(json.dumps({'unchanged': snapshots == after, 'after': after}, indent=2) + '\n')
