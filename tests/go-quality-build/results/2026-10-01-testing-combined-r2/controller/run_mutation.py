from pathlib import Path
import difflib, hashlib, json, os, shutil, subprocess, sys, tempfile, time

root=Path.cwd()
run_id,name,mutation,patch_file=sys.argv[1:5]
arc=root/"tests/go-quality-build/results"/run_id/name
patch=Path(patch_file).read_text()
destination=arc/"mutations"/mutation
destination.mkdir(parents=True,exist_ok=True)
(destination/"source.patch").write_text(patch)
def check(args,cwd):
 start=time.monotonic()
 env=dict(os.environ,GOCACHE="/private/tmp/go-quality-testing-cache",GOTOOLCHAIN="local")
 try:
  p=subprocess.run(["rtk","proxy",*args],cwd=cwd,env=env,text=True,capture_output=True,timeout=55)
  return dict(command=["rtk","proxy",*args],cwd=str(cwd),env={"GOCACHE":env["GOCACHE"],"GOTOOLCHAIN":"local"},stdout=p.stdout,stderr=p.stderr,exit_code=p.returncode,duration_seconds=round(time.monotonic()-start,3))
 except subprocess.TimeoutExpired as e:
  return dict(command=["rtk","proxy",*args],cwd=str(cwd),exit_code=None,timeout=True,stdout=str(e.stdout),stderr=str(e.stderr))
with tempfile.TemporaryDirectory(prefix="go-mutation-",dir="/private/tmp") as tmp:
 work=Path(tmp)/"module"; shutil.copytree(arc/"candidate",work)
 clean=check(["go","test","-count=1","-timeout=30s","./..."],work)
 applied=check(["git","apply",str(destination/"source.patch")],work)
 assert applied["exit_code"]==0,applied
 compiled=check(["go","test","-run","^$","-timeout=30s","./..."],work)
 result=check(["go","test","-count=1","-timeout=30s","./..."],work)
 record=dict(semantic_id=mutation,candidate_tests_only=True,controller_probes_absent=True,
  patch_sha256=hashlib.sha256(patch.encode()).hexdigest(),clean=clean,applied=applied,compiled=compiled,result=result,
  classification="pending manual assertion interpretation")
 if clean["exit_code"]!=0: record["classification"]="invalid clean baseline"
 elif compiled["exit_code"]!=0: record["classification"]="compile rejection, not behavioral detection"
 elif result["exit_code"]==0: record["classification"]="survived"
 elif result["exit_code"] is None: record["classification"]="timeout, not counted without interpretable assertion"
 else: record["classification"]="test failure; inspect named behavioral assertions before counting"
 (destination/"verification.json").write_text(json.dumps(record,indent=2)+"\n")
 print(mutation,record["classification"])
 print(result["stdout"][-7000:]); print(result["stderr"][-2000:])
