from pathlib import Path
import json, os, shutil, subprocess, tempfile, time, sys
arc=Path.cwd()/'tests/go-quality-build/results'/sys.argv[1]/sys.argv[2];tag=sys.argv[3];tasks=json.loads(Path(sys.argv[4]).read_text());records=[]
for task in tasks:
 with tempfile.TemporaryDirectory(prefix='go-focused-check-',dir='/private/tmp') as tmp:
  work=Path(tmp)/'module';shutil.copytree(arc/'candidate',work)
  if task.get('patch'):
   r=subprocess.run(['rtk','proxy','git','apply',str(Path(task['patch']).resolve())],cwd=work,capture_output=True,text=True);assert r.returncode==0,r.stderr
  cmd=['rtk','proxy',*task['args']];start=time.monotonic();r=subprocess.run(cmd,cwd=work,env=dict(os.environ,GOCACHE='/private/tmp/go-quality-testing-cache',GOTOOLCHAIN='local'),capture_output=True,text=True,timeout=55)
  records.append(dict(label=task['label'],command=cmd,cwd=str(work),env={'GOCACHE':'/private/tmp/go-quality-testing-cache','GOTOOLCHAIN':'local'},patch=task.get('patch'),stdout=r.stdout,stderr=r.stderr,exit_code=r.returncode,duration_seconds=time.monotonic()-start))
  print(task['label'],r.returncode, r.stdout.strip() if r.returncode else '')
(arc/(tag+'.json')).write_text(json.dumps(records,indent=2)+'\n')
