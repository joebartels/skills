import argparse
import hashlib
import json
import os
import re
import shutil
import subprocess
import time
from pathlib import Path

ROOT = Path('/Users/jb/.codex/worktrees/b633/skills')
OUT = Path('/private/tmp/go-isolation-promotion-review-20261001')
ENV = {'GOCACHE':'/private/tmp/go-quality-testing-cache', 'GOTOOLCHAIN':'local', 'GOPROXY':'off', 'GOSUMDB':'off', 'GOWORK':'off'}
parser = argparse.ArgumentParser()
parser.add_argument('--phase', choices=['ordinary','http-restricted','http-approved'], required=True)
args = parser.parse_args()
checks = []
path = OUT/(args.phase + '-checks.json')

def run(label, cwd, cmd, expected=0):
    full = ['rtk','proxy','env'] + [k+'='+v for k,v in ENV.items()] + cmd
    start = time.monotonic()
    try:
        r = subprocess.run(full, cwd=cwd, capture_output=True, text=True, timeout=60)
        record = {'label':label,'argv':full,'cwd':str(cwd),'env':ENV,'stdout':r.stdout,'stderr':r.stderr,'exit_code':r.returncode,'expected_exit_code':expected,'duration_seconds':round(time.monotonic()-start,3)}
    except subprocess.TimeoutExpired as e:
        record = {'label':label,'argv':full,'cwd':str(cwd),'env':ENV,'stdout':str(e.stdout),'stderr':str(e.stderr),'exit_code':None,'outer_timeout_seconds':60,'expected_exit_code':expected,'duration_seconds':round(time.monotonic()-start,3)}
    checks.append(record)
    path.write_text(json.dumps(checks,indent=2)+'\n')
    print(label, record['exit_code'], flush=True)
    return record

def copy(source, dest):
    if dest.exists():
        raise RuntimeError('Preserve prior attempt: '+str(dest))
    shutil.copytree(source,dest)
    return dest

run('toolchain', OUT, ['go','version'])
if args.phase == 'ordinary':
    cases = ['library-file-fixtures','cli-environment','worker-timing','not-isolation-work']
elif args.phase == 'http-restricted':
    cases = ['service-http-boundary']
else:
    cases = ['service-http-boundary']
for arm in ['baseline','skill-on']:
    archive = ROOT/('tests/go-quality-build/results/2026-10-01-test-isolation-'+arm)
    for case in cases:
        source = OUT/'reconstructed'/arm/case
        d = copy(source,OUT/'verification'/args.phase/arm/case)
        run(arm+':'+case+':clean',d,['go','test','-count=1','-timeout=30s','./...'])
        if args.phase == 'http-restricted':
            continue
        run(arm+':'+case+':vet',d,['go','vet','./...'])
        run(arm+':'+case+':gofmt',d,['gofmt','-l','.'])
        if case != 'not-isolation-work':
            run(arm+':'+case+':race-shuffle',d,['go','test','-race','-count=10','-shuffle=104729','-timeout=45s','./...'])
        if case == 'library-file-fixtures':
            selected = ['^TestStoreIndependentInstances$/^instances$/^alpha$']
            if arm == 'baseline':
                selected.append('^TestStoreIndependentInstances$/^instances$/^alpha$/^replacements$/^value-1$')
            else:
                selected.append('^TestStoreIndependentInstances$/^instances$/^alpha$/^shorter$')
            for index,pattern in enumerate(selected):
                run(arm+':file:focused-'+str(index),d,['go','test','-race','-count=1','-run',pattern,'-timeout=30s','./...'],expected=1 if arm=='baseline' else 0)
        for m in sorted((archive/case/'mutations').glob('*')):
            mutant = copy(source,OUT/'mutation-verification'/args.phase/arm/case/m.name)
            run(arm+':'+m.name+':patch',mutant,['git','apply',str(m/'source.patch')])
            run(arm+':'+m.name+':compile',mutant,['go','test','-run','^$','-timeout=30s','./...'])
            run(arm+':'+m.name+':signal',mutant,['go','test','-count=1','-timeout=30s','./...'],expected=1)
            if arm=='skill-on' and m.name=='store-shared-root':
                run('skill-on:shared-root:one-slot',mutant,['go','test','-count=1','-parallel=1','-run','^TestStoreIndependentInstances$','-timeout=30s','./...'],expected=1)
        before = {str(p.relative_to(d)):hashlib.sha256(p.read_bytes()).hexdigest() for p in d.rglob('*') if p.is_file()}
        original = {str(p.relative_to(source)):hashlib.sha256(p.read_bytes()).hexdigest() for p in source.rglob('*') if p.is_file()}
        checks.append({'label':arm+':'+case+':source-preserved','equal':before==original,'hashes':before})
        path.write_text(json.dumps(checks,indent=2)+'\n')

if args.phase=='ordinary':
    reference = ROOT/'docs/go-quality-build/drafts/go-test-isolation/references/isolation-patterns.md'
    code = re.search(r'```go\n(.*?)```',reference.read_text(),re.S).group(1)
    example = OUT/'example'
    example.mkdir()
    (example/'go.mod').write_text('module example.com/isolationillustration\n\ngo 1.22\n')
    (example/'example_test.go').write_text(code)
    run('exact-example:race',example,['go','test','-race','-count=1','-timeout=30s','./...'])
    run('exact-example:stdversion',example,['go','vet','-stdversion','./...'])
    mutant = copy(example,OUT/'example-premature-release')
    s=(mutant/'example_test.go').read_text()
    s=s.replace('go func() {','if err := file.Close(); err != nil { t.Fatal(err) }\n\tgo func() {',1)
    (mutant/'example_test.go').write_text(s)
    run('example-premature-release:compile',mutant,['go','test','-run','^$','-timeout=30s','./...'])
    run('example-premature-release:signal',mutant,['go','test','-count=1','-timeout=30s','./...'],expected=1)
    run('draft-quick-validate', ROOT,['python3','-B','/Users/jb/.codex/skills/.system/skill-creator/scripts/quick_validate.py','docs/go-quality-build/drafts/go-test-isolation'])
    run('repository-validate', ROOT,['python3','-B','scripts/validate.py'])
print(json.dumps({'phase':args.phase,'records':len(checks),'unexpected':[c['label'] for c in checks if ('exit_code' in c and c['exit_code'] != c['expected_exit_code']) or c.get('equal') is False]}),flush=True)
