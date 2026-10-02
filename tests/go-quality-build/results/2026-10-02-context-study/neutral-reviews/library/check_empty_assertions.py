from pathlib import Path
import json
import os
import shutil
import subprocess
import time

PACKET = Path('/private/tmp/go-context-outcome-20261002/library')
PATH = PACKET/'reviews/checks.json'
data = json.loads(PATH.read_text())
scratch = Path(data['scratch'])
env_changes = dict(GOWORK='off', GOTOOLCHAIN='local', GOPROXY='off', GOCACHE=str(scratch/'gocache'), GOMODCACHE=str(scratch/'gomodcache'))

def run(label, args, cwd):
    argv = ['rtk', 'proxy', 'go'] + args
    start = time.monotonic()
    p = subprocess.run(argv, cwd=cwd, env=dict(os.environ, **env_changes), text=True, capture_output=True, timeout=90)
    row = dict(label=label, argv=argv, cwd=str(cwd), environment_overrides=env_changes, status=p.returncode, stdout=p.stdout, stderr=p.stderr, elapsed_seconds=round(time.monotonic()-start, 3))
    data['checks'].append(row)
    PATH.write_text(json.dumps(data, indent=2)+'\n')
    print(label, 'status='+str(p.returncode), p.stdout+p.stderr, flush=True)

for n in (1,2):
    dst = scratch/f'candidate-{n}-spurious-empty-callback'
    shutil.copytree(PACKET/f'candidate-{n}', dst)
    path = dst/'process.go'
    source = path.read_text()
    old = '\tfor _, job := range jobs {\n'
    new = '\tif len(jobs) == 0 {\n\t\t_ = apply(ctx, 0)\n\t}\n'+old
    assert source.count(old)==1
    path.write_text(source.replace(old,new,1))
    run(f'candidate-{n}/spurious-empty-callback/compile', ['test','-run','^$','./...'],dst)
    run(f'candidate-{n}/spurious-empty-callback/author', ['test','-count=1','-timeout=30s','./...'],dst)
    shutil.copyfile(PACKET/'provided-contract-probes/contract_test.go',dst/'contract_controller_test.go')
    run(f'candidate-{n}/spurious-empty-callback/held', ['test','-run','^TestResultDecisionOrder$','-count=1','-timeout=30s','./...'],dst)
