import concurrent.futures
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import time

REPO = Path('/Users/jb/.codex/worktrees/b633/skills')
OUT = Path('/private/tmp/go-behavior-r2-promotion')
OUT.mkdir(parents=True, exist_ok=True)
ROOT = REPO / 'tests/go-quality-build/results'
RUNS = ['2026-10-01-behavior-tests-baseline', '2026-10-01-behavior-tests-skill-on', '2026-10-01-behavior-tests-r2', '2026-10-01-testing-combined', '2026-10-01-testing-combined-r2', '2026-10-01-testing-combined-final']
ENV = {'GOCACHE': str(OUT/'cache'), 'GOTOOLCHAIN': 'local'}
CHECKS = {'scope': RUNS, 'commands': [], 'integrity': [], 'matches': [], 'selection': [], 'limitations': []}

def sha(data):
    return hashlib.sha256(data).hexdigest()

def filemap(path):
    return {str(p.relative_to(path)): sha(p.read_bytes()) for p in sorted(path.rglob('*')) if p.is_file()}

def command(args, cwd=REPO, env=None, timeout=60):
    overrides = ENV if env is None else env
    start=time.monotonic()
    try:
        p=subprocess.run(['rtk','proxy',*args],cwd=cwd,env={**os.environ,**overrides},capture_output=True,text=True,timeout=timeout)
        item={'command':['rtk','proxy',*args], 'cwd':str(cwd), 'environment_overrides':overrides, 'stdout':p.stdout,'stderr':p.stderr,'exit_code':p.returncode,'duration_seconds':round(time.monotonic()-start,3)}
    except subprocess.TimeoutExpired as e:
        item={'command':['rtk','proxy',*args], 'cwd':str(cwd), 'environment_overrides':overrides, 'stdout':str(e.stdout),'stderr':str(e.stderr),'exit_code':None,'timeout_seconds':timeout,'duration_seconds':round(time.monotonic()-start,3)}
    CHECKS['commands'].append(item)
    return item

def save():
    (OUT/'checks.json').write_text(json.dumps(CHECKS,indent=2)+'\n')

def gitbytes(commit,path):
    r=command(['git','show',f'{commit}:{path}'],env={})
    assert r['exit_code']==0,r
    return r['stdout'].encode()

def task(dispatch):
    return dispatch.split('Complete this request:\n',1)[1].split('\n\nThe only optional',1)[0]

manifests={run:json.loads((ROOT/run/'manifest.json').read_text()) for run in RUNS}
canonical=(REPO/'docs/go-quality-build/README.md').read_text()
common=None
for run,m in manifests.items():
    rp=ROOT/run
    seal=rp/'checksums.sha256';failed=[];entries=seal.read_text().splitlines()
    for line in entries:
        expected,rel=line.split('  ',1)
        fp=rp/rel
        if not fp.is_file() or sha(fp.read_bytes())!=expected: failed.append(rel)
    sealhash=sha(seal.read_bytes())
    CHECKS['integrity'].append({'kind':'archive-seal','run':run,'entries':len(entries),'index_sha256':sealhash,'canonical_anchor_present':sealhash in canonical,'failed':failed})
    assert not failed
    for rel,expected in m.get('artifact_sha256',{}).items():
        assert sha((rp/rel).read_bytes())==expected,(run,rel)
    for rel,expected in m['skill_sha256'].items():
        assert sha((rp/'skills'/rel).read_bytes())==expected,(run,rel)
    common_map={k:v for k,v in m['skill_sha256'].items() if not k.startswith(('go-behavior-tests/','go-test-isolation/'))}
    if common is None: common=common_map
    assert common_map==common,(run,'common guidance mismatch')
    assert m['author_settings']==manifests[RUNS[0]]['author_settings'],(run,'settings mismatch')
    for c in m['cases']:
        cp=rp/c['id']; original=filemap(cp/'original');candidate=filemap(cp/'candidate')
        assert c['completed'],(run,c['id'])
        assert original==c['input_sha256'],(run,c['id'],'input mismatch')
        assert candidate==c['source_sha256'],(run,c['id'],'candidate mismatch')
        assert sha((cp/'dispatch.txt').read_bytes())==c['dispatch_sha256']
        for rel,expected in c.get('artifact_sha256',{}).items():
            assert sha((cp/rel).read_bytes())==expected,(run,c['id'],rel)
        # Every original file is checked against its frozen preparation commit.
        suite='go-behavior-tests' if 'behavior-tests-' in run else 'testing-combined'
        fixture=c.get('fixture_id',c['id'])
        for rel,expected in original.items():
            p=f'tests/go-quality-build/{suite}/evals/files/{fixture}/{rel}'
            assert sha(gitbytes(m['fixture_commit'],p))==expected,(run,c['id'],p)
        reconstructed=OUT/'reconstructed'/run/c['id']
        if reconstructed.exists(): shutil.rmtree(reconstructed)
        shutil.copytree(cp/'original',reconstructed)
        applied=command(['git','apply',str(cp/'source.patch')],reconstructed,env={})
        assert applied['exit_code']==0
        assert filemap(reconstructed)==candidate,(run,c['id'],'reconstruction mismatch')
        CHECKS['integrity'].append({'kind':'candidate-reconstruction','run':run,'case':c['id'],'files':len(candidate),'reconstructed_path':str(reconstructed),'matched':True})
        if 'behavior-tests-' in run:
            baseline=ROOT/RUNS[0]/fixture
            assert task((cp/'dispatch.txt').read_text())==task((baseline/'dispatch.txt').read_text()),(run,c['id'],'task mismatch')
            assert (cp/'prompt.txt').read_bytes()==(baseline/'prompt.txt').read_bytes(),(run,c['id'],'prompt mismatch')
            assert original==filemap(baseline/'original')
        else:
            baseline=ROOT/'2026-10-01-testing-combined'/'architecture-only'
            assert task((cp/'dispatch.txt').read_text())==task((baseline/'dispatch.txt').read_text()),(run,c['id'],'task mismatch')
            assert (cp/'prompt.txt').read_bytes()==(baseline/'prompt.txt').read_bytes(),(run,c['id'],'prompt mismatch')
            assert original==filemap(baseline/'original')
        for rel,expected in c.get('guidance_sha256',{}).items():
            assert sha((rp/'skills'/rel).read_bytes())==expected
        CHECKS['matches'].append({'run':run,'case':c['id'],'task':True,'prompt':True,'input':True,'common_guidance':True,'inherited_settings':True})
        if run in ['2026-10-01-behavior-tests-r2','2026-10-01-testing-combined-final']:
            sel=json.loads((cp/'selection.json').read_text())
            # Preserve raw heterogeneous selection schemas, rather than inferring routing.
            CHECKS['selection'].append({'run':run,'case':c['id'],'raw':sel})

for rel in ['SKILL.md','references/behavior-observations.md']:
    source=gitbytes('7812d4be41c8112aaf2aa8049e63b4a0ad57228e',f'docs/go-quality-build/drafts/go-behavior-tests/{rel}')
    expected=sha(source)
    paths=[REPO/'docs/go-quality-build/drafts/go-behavior-tests'/rel,ROOT/'2026-10-01-behavior-tests-r2/skills/go-behavior-tests'/rel,ROOT/'2026-10-01-testing-combined-final/skills/go-behavior-tests'/rel]
    assert all(p.read_bytes()==source for p in paths)
    CHECKS['integrity'].append({'kind':'exact-guidance','source_commit':'7812d4be41c8112aaf2aa8049e63b4a0ad57228e','path':rel,'sha256':expected,'matching_paths':[str(p) for p in paths]})

old=(ROOT/'2026-10-01-behavior-tests-skill-on/skills/go-behavior-tests/references/behavior-observations.md').read_text()
new=(REPO/'docs/go-quality-build/drafts/go-behavior-tests/references/behavior-observations.md').read_text()
assert re.findall(r'```go\n(.*?)```',old,re.S)==re.findall(r'```go\n(.*?)```',new,re.S)
CHECKS['integrity'].append({'kind':'unchanged-go-illustration','matched':True,'blocks':len(re.findall(r'```go\n(.*?)```',new,re.S))})

for rp in [ROOT/'2026-10-01-behavior-tests-r2',ROOT/'2026-10-01-testing-combined-final']:
    sealed=command(['git','diff','--exit-code','b187873','--',str(rp.relative_to(REPO))],env={})
    assert sealed['exit_code']==0

save()
print(json.dumps({'seals':[x for x in CHECKS['integrity'] if x['kind']=='archive-seal'],'candidate_reconstructions':len([x for x in CHECKS['integrity'] if x['kind']=='candidate-reconstruction']),'matched_trials':len(CHECKS['matches']),'guidance':[x for x in CHECKS['integrity'] if x['kind']=='exact-guidance'],'commands':len(CHECKS['commands'])},indent=2),flush=True)

def clean(run,case):
    p=OUT/'reconstructed'/run/case
    items=[]
    for args in [['go','test','-count=1','-timeout=45s','./...'],['go','vet','./...'],['gofmt','-l','.']]:
        r=command(args,p,timeout=60);items.append(r)
    return {'run':run,'case':case,'exit_codes':[x['exit_code'] for x in items],'format_stdout':items[-1]['stdout']}

targets=[(run,c['id']) for run in ['2026-10-01-behavior-tests-r2','2026-10-01-testing-combined-final'] for c in manifests[run]['cases']]
CHECKS['fresh_clean']=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
    jobs={pool.submit(clean,*t):t for t in targets}
    for future in concurrent.futures.as_completed(jobs):
        result=future.result();CHECKS['fresh_clean'].append(result);save();print(json.dumps(result),flush=True)
save()
