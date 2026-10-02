from pathlib import Path
import hashlib,json,os,shutil,subprocess,time
ROOT=Path('/Users/jb/.codex/worktrees/b633/skills')
OUT=Path('/private/tmp/go-behavior-promotion-review')
PIN='716e779b2470021ae64602a197fbafb979eec0b5'
records=[]
env=dict(os.environ,GOCACHE='/private/tmp/go-quality-testing-cache',GOTOOLCHAIN='local',GOPROXY='off')
def run(args,cwd):
 start=time.monotonic()
 try:
  p=subprocess.run(['rtk','proxy',*args],cwd=cwd,env=env,text=True,capture_output=True,timeout=55)
  r=dict(command=['rtk','proxy',*args],cwd=str(cwd),env={k:env[k] for k in ['GOCACHE','GOTOOLCHAIN','GOPROXY']},stdout=p.stdout,stderr=p.stderr,exit_code=p.returncode,duration_seconds=round(time.monotonic()-start,3))
 except subprocess.TimeoutExpired as e:
  r=dict(command=['rtk','proxy',*args],cwd=str(cwd),exit_code=None,timeout=True,stdout=str(e.stdout),stderr=str(e.stderr))
 records.append(r)
 (OUT/'probe-checks.json').write_text(json.dumps(records,indent=2)+'\n')
 return r
for arm in ['baseline','skill-on']:
 for case in ['library-codec','cli-partial-failure','service-publication','worker-host']:
  target=OUT/'probed'/arm/case
  shutil.copytree(OUT/'reconstructed'/arm/case,target)
  probe_root=ROOT/f'tests/go-quality-build/go-behavior-tests/evals/controller-probes/{case}'
  for file in probe_root.rglob('*'):
   if not file.is_file(): continue
   source=run(['git','show',f'{PIN}:{file.relative_to(ROOT)}'],ROOT)
   assert source['exit_code']==0 and source['stdout'].encode()==file.read_bytes()
   records.append(dict(kind='frozen-probe-byte-match',path=str(file.relative_to(ROOT)),sha256=hashlib.sha256(file.read_bytes()).hexdigest(),fixture_commit=PIN))
   shutil.copy2(file,target/('controller_'+file.name))
  r=run(['go','test','-count=1','-timeout=30s','./...'],target)
  print(arm,case,r['exit_code'],r['stdout'][-500:],r['stderr'][-500:],flush=True)
(OUT/'probe-checks.json').write_text(json.dumps(records,indent=2)+'\n')
