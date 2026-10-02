from pathlib import Path
import json, os, shutil, subprocess, sys, time
ROOT=Path('/Users/jb/.codex/worktrees/b633/skills')
OUT=Path('/private/tmp/go-behavior-promotion-review')
mode=sys.argv[1]
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
 (OUT/f'{mode}-checks.json').write_text(json.dumps(records,indent=2)+'\n')
 return r
if mode in ['clean','service','service-approved']:
 for arm in ['baseline','skill-on']:
  for case in ['library-codec','cli-partial-failure','service-publication','worker-host','not-behavior-work']:
   if (case=='service-publication') != (mode!='clean'): continue
   source=OUT/'reconstructed'/arm/case
   commands=[['go','test','-count=1','-timeout=30s','./...'],['go','vet','./...'],['gofmt','-l','.']]
   if case in ['worker-host','service-publication']:
    commands.extend([['go','test','-race','-count=1','-timeout=30s','./...'],['go','test','-shuffle=on','-count=10','-timeout=30s','./...']])
   for cmd in commands:
    r=run(cmd,source)
    print(arm,case,' '.join(cmd),r['exit_code'],(r['stdout']+r['stderr'])[-1000:],flush=True)
elif mode in ['mutations','service-mutations','service-mutations-approved']:
 for arm in ['baseline','skill-on']:
  archive=ROOT/f'tests/go-quality-build/results/2026-10-01-behavior-tests-{arm}'
  for arc in sorted(archive.glob('*/mutations/*')):
   case=arc.parent.parent.name
   if (case=='service-publication') != (mode!='mutations'): continue
   target=OUT/'mutated'/mode/arm/case/arc.name
   shutil.copytree(OUT/'reconstructed'/arm/case,target)
   applied=run(['git','apply',str(arc/'source.patch')],target)
   if applied['exit_code']!=0:
    print(arm,case,arc.name,'PATCH FAILED',applied,flush=True); continue
   compiled=run(['go','test','-run','^$','-timeout=30s','./...'],target)
   r=run(['go','test','-count=1','-timeout=30s','./...'],target)
   print(arm,case,arc.name,'compiled',compiled['exit_code'],'suite',r['exit_code'],(r['stdout']+r['stderr'])[-2500:],flush=True)
else:
 raise SystemExit('unknown mode')
