from pathlib import Path
import hashlib
import json
import os
import shutil
import subprocess
import tempfile
import time

PACKET = Path('/private/tmp/go-context-outcome-20261002/library')
SCRATCH = Path(tempfile.mkdtemp(prefix='go-library-neutral-review-', dir='/private/tmp'))
OUT = PACKET / 'reviews'
OUT.mkdir(exist_ok=True)
RECORDS = []
ENV = dict(os.environ, GOWORK='off', GOTOOLCHAIN='local', GOPROXY='off', GOCACHE=str(SCRATCH / 'gocache'), GOMODCACHE=str(SCRATCH / 'gomodcache'))

def run(label, argv, cwd, extra_env=None):
    env = dict(ENV, **(extra_env or {}))
    start = time.monotonic()
    p = subprocess.run(argv, cwd=cwd, env=env, text=True, capture_output=True, timeout=90)
    row = dict(label=label, argv=argv, cwd=str(cwd), environment_overrides={k: env[k] for k in ('GOWORK', 'GOTOOLCHAIN', 'GOPROXY', 'GOCACHE', 'GOMODCACHE')}, status=p.returncode, stdout=p.stdout, stderr=p.stderr, elapsed_seconds=round(time.monotonic()-start, 3))
    RECORDS.append(row)
    (OUT / 'checks.json').write_text(json.dumps(dict(scratch=str(SCRATCH), checks=RECORDS), indent=2)+'\n')
    print(label, 'status='+str(p.returncode), 'elapsed='+str(row['elapsed_seconds']), flush=True)
    if p.returncode:
        print((p.stdout+p.stderr)[:2200], flush=True)
    return row

def copy_candidate(n, suffix):
    dst = SCRATCH / f'candidate-{n}-{suffix}'
    shutil.copytree(PACKET / f'candidate-{n}', dst)
    return dst

run('toolchain-version', ['rtk', 'proxy', 'go', 'version'], PACKET)
original = SCRATCH / 'original'
shutil.copytree(PACKET / 'original', original)
run('original-preserved-baseline', ['rtk', 'proxy', 'go', 'test', '-count=1', '-timeout=30s', './...'], original)

for n in (1, 2):
    author = copy_candidate(n, 'author')
    run(f'candidate-{n}/standalone-cold-build', ['rtk', 'proxy', 'go', 'build', '-mod=readonly', './...'], author,
        dict(GOCACHE=str(SCRATCH / f'candidate-{n}-cold-gocache'), GOMODCACHE=str(SCRATCH / f'candidate-{n}-cold-gomodcache')))
    run(f'candidate-{n}/author', ['rtk', 'proxy', 'go', 'test', '-count=1', '-timeout=30s', './...'], author)
    run(f'candidate-{n}/vet', ['rtk', 'proxy', 'go', 'vet', './...'], author)
    run(f'candidate-{n}/format', ['rtk', 'proxy', 'gofmt', '-d', '.'], author)
    run(f'candidate-{n}/module-graph', ['rtk', 'proxy', 'go', 'list', '-m', 'all'], author)
    held = copy_candidate(n, 'contract')
    shutil.copyfile(PACKET / 'provided-contract-probes/contract_test.go', held / 'contract_controller_test.go')
    run(f'candidate-{n}/provided-contract', ['rtk', 'proxy', 'go', 'test', '-count=1', '-timeout=30s', './...'], held)
    run(f'candidate-{n}/provided-contract-race', ['rtk', 'proxy', 'go', 'test', '-race', '-shuffle=on', '-count=3', '-timeout=60s', './...'], held)
    for mutation in json.loads((PACKET / f'provided-checks-{n}.json').read_text())['frozen_mutation_sensitivity']:
        dst = copy_candidate(n, mutation['id'])
        for sub in mutation['substitutions']:
            path = dst / sub['file']
            source = path.read_text()
            assert sub['old'] in source
            path.write_text(source.replace(sub['old'], sub['new']))
        run(f'candidate-{n}/{mutation["id"]}/compile', ['rtk', 'proxy', 'go', 'test', '-run', '^$', './...'], dst)
        run(f'candidate-{n}/{mutation["id"]}/author', ['rtk', 'proxy', 'go', 'test', '-count=1', '-timeout=30s', './...'], dst)
        shutil.copyfile(PACKET / 'provided-contract-probes/contract_test.go', dst / 'contract_controller_test.go')
        name = 'TestCauseReturnedByCallback' if mutation['id']=='unsafe-cause-equality' else 'TestCustomCauseAndIndependentFailure'
        run(f'candidate-{n}/{mutation["id"]}/held', ['rtk', 'proxy', 'go', 'test', '-run', '^'+name+'$', '-count=1', '-timeout=30s', './...'], dst)

    dst = copy_candidate(n, 'remove-entry-cancellation-check')
    source = (dst / 'process.go').read_text()
    block = '\tif err := cancellationError(ctx); err != nil {\n\t\treturn 0, err\n\t}\n' if n==1 else '\tif ctx.Err() != nil {\n\t\treturn 0, cancellationError(ctx)\n\t}\n'
    assert source.count(block)==1
    (dst / 'process.go').write_text(source.replace(block, '', 1))
    run(f'candidate-{n}/remove-entry-cancellation-check/compile', ['rtk', 'proxy', 'go', 'test', '-run', '^$', './...'], dst)
    run(f'candidate-{n}/remove-entry-cancellation-check/author', ['rtk', 'proxy', 'go', 'test', '-count=1', '-timeout=30s', './...'], dst)
    shutil.copyfile(PACKET / 'provided-contract-probes/contract_test.go', dst / 'contract_controller_test.go')
    run(f'candidate-{n}/remove-entry-cancellation-check/held', ['rtk', 'proxy', 'go', 'test', '-run', '^TestCanceledBeforeStart$', '-count=1', '-timeout=30s', './...'], dst)

snapshots = {str(p.relative_to(PACKET)): hashlib.sha256(p.read_bytes()).hexdigest() for folder in ('original','candidate-1','candidate-2') for p in sorted((PACKET/folder).rglob('*')) if p.is_file()}
(OUT / 'source-hashes.json').write_text(json.dumps(snapshots, indent=2)+'\n')
print('Scratch:', SCRATCH, flush=True)
