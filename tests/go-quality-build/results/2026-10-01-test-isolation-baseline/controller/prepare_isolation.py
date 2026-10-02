from pathlib import Path
import json, os, subprocess, time, tempfile, shutil, hashlib
root=Path.cwd()
suite=root/"tests/go-quality-build/go-test-isolation/evals"
cache=Path("/private/tmp/go-quality-testing-cache"); cache.mkdir(exist_ok=True)
env=dict(os.environ,GOCACHE=str(cache),GOTOOLCHAIN="local")
def check(args,cwd):
 start=time.monotonic()
 try:
  p=subprocess.run(["rtk","proxy",*args],cwd=cwd,env=env,text=True,capture_output=True,timeout=55)
  return dict(command=["rtk","proxy",*args],cwd=str(cwd),env={"GOCACHE":str(cache),"GOTOOLCHAIN":"local"},stdout=p.stdout,stderr=p.stderr,exit_code=p.returncode,duration_seconds=round(time.monotonic()-start,3))
 except subprocess.TimeoutExpired as e:
  return dict(command=["rtk","proxy",*args],cwd=str(cwd),exit_code=None,timeout=True,stdout=str(e.stdout),stderr=str(e.stderr))
def hashes(path):
 return {str(p.relative_to(path)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(path.rglob("*")) if p.is_file()}
data=json.loads((suite/"evals.json").read_text())
results={}
for case in data["evals"]:
 name=case["id"]; fixture=suite/"files"/name
 assert sorted(case["files"])==sorted(str(p.relative_to(suite)) for p in fixture.rglob("*") if p.is_file())
 out=[check(a,fixture) for a in [["go","test","-count=1","-timeout=30s","./..."],["go","vet","./..."],["gofmt","-l","."]]]
 assert all(x["exit_code"]==0 for x in out),out
 assert not out[-1]["stdout"],out[-1]
 probes=suite/"controller-probes"/name
 if probes.exists():
  with tempfile.TemporaryDirectory(prefix="go-testing-input-") as tmp:
   copied=Path(tmp)/name; shutil.copytree(fixture,copied)
   for p in probes.rglob("*"):
    if p.is_file(): shutil.copy2(p,copied/p.name)
   out.append(check(["go","test","-count=1","-timeout=30s","./..."],copied))
 results[name]={"input_hashes":hashes(fixture),"checks":out}
 print(name,"old tests/vet/format PASS; controller contract probes",out[-1]["exit_code"] if probes.exists() else "not applicable")
run=root/"tests/go-quality-build/results/2026-10-01-test-isolation-baseline"
run.mkdir(parents=True,exist_ok=True)
(run/"preparation.json").write_text(json.dumps({"cases":results,"controller_probe_hashes":hashes(suite/"controller-probes")},indent=2)+"\n")
