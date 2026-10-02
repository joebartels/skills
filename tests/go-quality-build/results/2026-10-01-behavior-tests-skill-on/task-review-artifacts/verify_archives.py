from pathlib import Path
import hashlib, json, os, re, shutil, subprocess, time

ROOT = Path('/Users/jb/.codex/worktrees/b633/skills')
OUT = Path('/private/tmp/go-behavior-promotion-review')
RUNS = {arm: ROOT / f'tests/go-quality-build/results/2026-10-01-behavior-tests-{arm}' for arm in ['baseline', 'skill-on']}
records = []
summary = {}
def hashes(p):
    return {str(f.relative_to(p)): hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted(p.rglob('*')) if f.is_file()}
def check(args, cwd=ROOT, env=None):
    effective = dict(os.environ, GOCACHE='/private/tmp/go-quality-testing-cache', GOTOOLCHAIN='local', GOPROXY='off')
    if env: effective.update(env)
    start = time.monotonic()
    try:
        p = subprocess.run(['rtk','proxy',*args], cwd=cwd, env=effective, capture_output=True, text=True, timeout=55)
        r = dict(command=['rtk','proxy',*args], cwd=str(cwd), env={k:effective[k] for k in ['GOCACHE','GOTOOLCHAIN','GOPROXY']}, exit_code=p.returncode, stdout=p.stdout, stderr=p.stderr, duration_seconds=round(time.monotonic()-start,3))
    except subprocess.TimeoutExpired as e:
        r = dict(command=['rtk','proxy',*args], cwd=str(cwd), exit_code=None, timeout=True, stdout=str(e.stdout), stderr=str(e.stderr))
    records.append(r)
    (OUT/'archive-checks.json').write_text(json.dumps(dict(summary=summary, checks=records),indent=2)+'\n')
    return r
def require(ok, message):
    records.append(dict(kind='assertion', description=message, result='pass' if ok else 'fail'))
    assert ok, message

check(['git','rev-parse','HEAD'])
checkpoint = check(['git','rev-parse','cf2a5b2'])['stdout'].strip()
draft_commit='3bf12e8ef0a4c9740e53e8fe5b39a1c08af94a2b'
manifests={a:json.loads((p/'manifest.json').read_text()) for a,p in RUNS.items()}
for arm,p in RUNS.items():
    declared={}
    for line in (p/'checksums.sha256').read_text().splitlines():
        digest,name=line.split('  ',1)
        declared[name]=digest
    actual=hashes(p); actual.pop('checksums.sha256')
    require(declared==actual,f'{arm}: checksum index equals all {len(actual)} archive files except itself')
    summary[arm]={'artifact_count':len(actual),'checksum_index_sha256':hashlib.sha256((p/'checksums.sha256').read_bytes()).hexdigest()}
    checkpoint_bytes=check(['git','show',f'{checkpoint}:{p.relative_to(ROOT)}/checksums.sha256'])['stdout'].encode()
    require(checkpoint_bytes==(p/'checksums.sha256').read_bytes(),f'{arm}: checksum index equals evidence checkpoint')
    m=manifests[arm]
    for name,digest in m['skill_sha256'].items():
        require(hashlib.sha256((p/'skills'/name).read_bytes()).hexdigest()==digest,f'{arm}: skill snapshot {name}')
    if 'artifact_sha256' in m:
        for name,digest in m['artifact_sha256'].items():
            require(hashlib.sha256((p/name).read_bytes()).hexdigest()==digest,f'{arm}: manifest artifact {name}')
    for c in m['cases']:
        arc=p/c['id']
        require(hashes(arc/'original')==c['input_sha256'],f'{arm}/{c["id"]}: exact original hashes')
        require(hashes(arc/'candidate')==c['source_sha256'],f'{arm}/{c["id"]}: exact candidate hashes')
        require(hashlib.sha256((arc/'dispatch.txt').read_bytes()).hexdigest()==c['dispatch_sha256'],f'{arm}/{c["id"]}: dispatch hash')
        for name,digest in c['artifact_sha256'].items():
            require(hashlib.sha256((arc/name).read_bytes()).hexdigest()==digest,f'{arm}/{c["id"]}: artifact {name}')
        target=OUT/'reconstructed'/arm/c['id']
        shutil.copytree(arc/'original',target)
        result=check(['git','apply',str(arc/'source.patch')],target)
        require(result['exit_code']==0,f'{arm}/{c["id"]}: patch applies')
        require(hashes(target)==hashes(arc/'candidate'),f'{arm}/{c["id"]}: reconstruction byte equality')
        for name,digest in c['input_sha256'].items():
            fixture=ROOT/f'tests/go-quality-build/go-behavior-tests/evals/files/{c["id"]}/{name}'
            original_bytes=check(['git','show',f'{m["fixture_commit"]}:{fixture.relative_to(ROOT)}'])['stdout'].encode()
            require(hashlib.sha256(original_bytes).hexdigest()==digest,f'{arm}/{c["id"]}: frozen commit input {name}')
    print(arm,summary[arm],flush=True)

b,s=manifests['baseline'],manifests['skill-on']
require(b['author_settings']==s['author_settings'],'author settings exactly equal as exposed')
require(b['review_settings']==s['review_settings'],'review settings exactly equal as exposed')
require(b['host']==s['host'],'recorded host settings exactly equal')
for name,digest in b['skill_sha256'].items():
    require(s['skill_sha256'][name]==digest,f'matched existing guidance {name}')
for bc,sc in zip(b['cases'],s['cases']):
    require(bc['id']==sc['id'] and bc['input_sha256']==sc['input_sha256'],f'matched inputs {bc["id"]}')
    require((RUNS['baseline']/bc['id']/'prompt.txt').read_bytes()==(RUNS['skill-on']/sc['id']/'prompt.txt').read_bytes(),f'matched task prompt {bc["id"]}')
    def normalize_wrapper(arm,c):
        text=(RUNS[arm]/c['id']/'dispatch.txt').read_text()
        text=re.sub(r'/private/tmp/go-testing-author-[^/\s]+','/private/tmp/AUTHOR',text)
        text=re.sub(r'^- go-behavior-tests:.*\n','',text,flags=re.M)
        text=text.replace('The two new testing-writing skills are unavailable in this task.','The other new testing-writing skill is unavailable in this task.')
        return text
    require(normalize_wrapper('baseline',bc)==normalize_wrapper('skill-on',sc),f'matched author wrapper apart from paths/candidate availability {bc["id"]}')
for name,digest in s['candidate_sha256'].items():
    path=ROOT/'docs/go-quality-build/drafts/go-behavior-tests'/name
    committed=check(['git','show',f'{draft_commit}:{path.relative_to(ROOT)}'])['stdout'].encode()
    require(hashlib.sha256(committed).hexdigest()==digest,f'draft commit hash {name}')
    require(path.read_bytes()==committed==(RUNS['skill-on']/'skills/go-behavior-tests'/name).read_bytes(),f'draft/current/evaluated snapshot byte equality {name}')
check(['go','version'])
check(['go','env','GOVERSION','GOOS','GOARCH','GOMOD','GOWORK'])
summary['reconstructed_modules']=10
summary['source_file_hash_count']=sum(len(c['source_sha256']) for m in manifests.values() for c in m['cases'])
(OUT/'archive-checks.json').write_text(json.dumps(dict(summary=summary, checks=records),indent=2)+'\n')
print('summary',json.dumps(summary),flush=True)
