from pathlib import Path
import json,os,shutil,subprocess,tempfile,time,sys
root=Path.cwd();run=root/'tests/go-quality-build/results'/sys.argv[1];arc=run/sys.argv[2];records=[]
home=Path('/private/tmp/go-testing-default-home');home.mkdir(exist_ok=True)
def check(args,cwd,env):
 start=time.monotonic()
 try:
  p=subprocess.run(['rtk','proxy',*args],cwd=cwd,env=env,capture_output=True,text=True,timeout=55)
  r=dict(command=['rtk','proxy',*args],cwd=str(cwd),env={'GOCACHE':env.get('GOCACHE','<unset>'),'HOME':env.get('HOME','<unset>'),'GOTOOLCHAIN':env.get('GOTOOLCHAIN')},stdout=p.stdout,stderr=p.stderr,exit_code=p.returncode,duration_seconds=round(time.monotonic()-start,3))
 except subprocess.TimeoutExpired as e:r=dict(command=args,cwd=str(cwd),exit_code=None,timeout=True,stdout=str(e.stdout),stderr=str(e.stderr))
 records.append(r);return r
with tempfile.TemporaryDirectory(prefix='go-default-cache-',dir='/private/tmp') as tmp:
 work=Path(tmp)/'module';shutil.copytree(arc/'candidate',work)
 buildenv=dict(os.environ,GOCACHE='/private/tmp/go-quality-testing-cache',GOTOOLCHAIN='local')
 packages=check(['go','list','-f','{{if .TestGoFiles}}{{.ImportPath}}{{else if .XTestGoFiles}}{{.ImportPath}}{{end}}','./...'],work,buildenv);assert packages['exit_code']==0
 for n,package in enumerate(packages['stdout'].split()):
  exe=Path(tmp)/('tests-'+str(n));compiled=check(['go','test','-c','-o',str(exe),package],work,buildenv);assert compiled['exit_code']==0
  env=dict(buildenv,HOME=str(home));env.pop('GOCACHE',None)
  info=check(['go','list','-f','{{.Dir}}',package],work,buildenv);assert info['exit_code']==0
  check([str(exe),'-test.count=1','-test.timeout=30s'],Path(info['stdout'].strip()),env)
(arc/'default-cache-verification.json').write_text(json.dumps({'scope':'compile candidate tests with controller cache; run unchanged test binaries with GOCACHE unset and valid disposable HOME, exercising child builder defaults without outer go-test cache setup; no injected assertions','checks':records},indent=2)+'\n')
print(sys.argv[2],[(r['command'][2],r['exit_code']) for r in records])
