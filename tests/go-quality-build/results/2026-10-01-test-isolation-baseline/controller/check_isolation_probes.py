from pathlib import Path
import json, os, shutil, subprocess, tempfile, time

root=Path.cwd()
suite=root/'tests/go-quality-build/go-test-isolation/evals'
run=root/'tests/go-quality-build/results/2026-10-01-test-isolation-baseline'
env=dict(os.environ,GOCACHE='/private/tmp/go-quality-testing-cache',GOTOOLCHAIN='local')
checks=[]
for name in ['library-file-fixtures','cli-environment','service-http-boundary','worker-timing']:
 with tempfile.TemporaryDirectory(prefix='go-isolation-probe-',dir='/private/tmp') as tmp:
  work=Path(tmp)/'module';shutil.copytree(suite/'files'/name,work)
  for p in (suite/'controller-probes'/name).glob('*.go'):shutil.copyfile(p,work/p.name)
  args=['rtk','proxy','go','test','-race','-count=10','-shuffle=on','-timeout=30s','./...']
  start=time.monotonic();p=subprocess.run(args,cwd=work,env=env,text=True,capture_output=True,timeout=50)
  checks.append(dict(case=name,command=args,cwd=str(work),env={'GOCACHE':env['GOCACHE'],'GOTOOLCHAIN':'local'},stdout=p.stdout,stderr=p.stderr,exit_code=p.returncode,duration_seconds=time.monotonic()-start,execution_permission='automatically approved local loopback evaluation'))
  print(name,p.returncode,p.stdout.strip())
assert all(c['exit_code']==0 for c in checks),checks
(run/'probe-repeat-verification.json').write_text(json.dumps(checks,indent=2)+'\n')
