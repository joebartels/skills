from pathlib import Path
import json, os, re, subprocess, sys, tempfile, shutil, time

run_id, case = sys.argv[1:3]
arc = Path.cwd()/'tests/go-quality-build/results'/run_id/case
results=[]
with tempfile.TemporaryDirectory(prefix='go-isolated-tests-',dir='/private/tmp') as tmp:
    work=Path(tmp)/'module';shutil.copytree(arc/'candidate',work)
    names=sorted({name for p in work.rglob('*_test.go') for name in re.findall(r'^func (Test\w+)\(t \*testing.T\)',p.read_text(),re.M)})
    for name in names:
        cmd=['rtk','proxy','go','test','-race','-count=1','-timeout=30s','-run','^'+name+'$','./...']
        env=dict(os.environ,GOCACHE='/private/tmp/go-quality-testing-cache',GOTOOLCHAIN='local')
        start=time.monotonic()
        p=subprocess.run(cmd,cwd=work,env=env,text=True,capture_output=True,timeout=55)
        results.append(dict(test=name,command=cmd,cwd=str(work),env={'GOCACHE':env['GOCACHE'],'GOTOOLCHAIN':'local'},stdout=p.stdout,stderr=p.stderr,exit_code=p.returncode,duration_seconds=round(time.monotonic()-start,3)))
(arc/'individual-verification.json').write_text(json.dumps(results,indent=2)+'\n')
print(case,[(r['test'],r['exit_code']) for r in results])
