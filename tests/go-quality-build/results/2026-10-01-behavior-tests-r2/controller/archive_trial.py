from pathlib import Path
import difflib, hashlib, json, os, shutil, subprocess, sys, tempfile, time

root=Path.cwd()
suite_name,run_id,name=sys.argv[1:4]
suite=root/"tests/go-quality-build"/suite_name/"evals"
run=root/"tests/go-quality-build/results"/run_id
manifest=json.loads((run/"manifest.json").read_text())
entry=next(c for c in manifest["cases"] if c["id"]==name)
work=Path(entry["workdir"]); report=Path(entry["reportdir"]); arc=run/name
def hashes(path):
 return {str(p.relative_to(path)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(path.rglob("*")) if p.is_file()}
def check(args,cwd):
 start=time.monotonic()
 env=dict(os.environ,GOCACHE="/private/tmp/go-quality-testing-cache",GOTOOLCHAIN="local")
 try:
  p=subprocess.run(["rtk","proxy",*args],cwd=cwd,env=env,text=True,capture_output=True,timeout=55)
  return dict(command=["rtk","proxy",*args],cwd=str(cwd),env={"GOCACHE":env["GOCACHE"],"GOTOOLCHAIN":"local"},stdout=p.stdout,stderr=p.stderr,exit_code=p.returncode,duration_seconds=round(time.monotonic()-start,3))
 except subprocess.TimeoutExpired as e:
  return dict(command=["rtk","proxy",*args],cwd=str(cwd),exit_code=None,timeout=True,stdout=str(e.stdout),stderr=str(e.stderr))
shutil.copytree(report,arc/"author-artifacts",dirs_exist_ok=True)
for filename in ["author-report.md","selection.json","checks.json"]:
 if (report/filename).exists(): shutil.copy2(report/filename,arc/filename)
for filename in ["author-report.md","selection.json"]: assert (arc/filename).exists(),filename
original=suite/"files"/entry.get("fixture_id",name)
shutil.copytree(original,arc/"original",dirs_exist_ok=True)
shutil.copytree(work,arc/"candidate",dirs_exist_ok=True)
names=sorted(set(hashes(original))|set(hashes(work)))
patch=""
for filename in names:
 a=original/filename; b=work/filename
 left=a.read_text().splitlines(keepends=True) if a.exists() else []
 right=b.read_text().splitlines(keepends=True) if b.exists() else []
 patch+="".join(difflib.unified_diff(left,right,fromfile="a/"+filename if a.exists() else "/dev/null",tofile="b/"+filename if b.exists() else "/dev/null"))
(arc/"source.patch").write_text(patch)
checks=[]
with tempfile.TemporaryDirectory(prefix="go-testing-reconstruct-",dir="/private/tmp") as tmp:
 reconstructed=Path(tmp)/"module"; shutil.copytree(original,reconstructed)
 if patch:
  applied=check(["git","apply",str(arc/"source.patch")],reconstructed)
  checks.append(applied); assert applied["exit_code"]==0,applied
 assert hashes(reconstructed)==hashes(work),"reconstruction differs"
 checks.append(dict(kind="patch-reconstruction",result="byte-for-byte",file_count=len(hashes(work))))
 for args in [["go","test","-count=1","-timeout=30s","./..."],["go","vet","./..."],["gofmt","-l","."]]:
  checks.append(check(args,reconstructed))
 if entry.get("isolation_relevant",False) or name in ["worker-host","worker-timing","service-publication","service-http-boundary","library-file-fixtures","cli-environment"]:
  checks.append(check(["go","test","-race","-count=1","-timeout=30s","./..."],reconstructed))
  checks.append(check(["go","test","-shuffle=on","-count=10","-timeout=30s","./..."],reconstructed))
 probes=suite/"controller-probes"/entry.get("fixture_id",name)
 if probes.exists():
  for p in probes.rglob("*"):
   if p.is_file(): shutil.copy2(p,reconstructed/("controller_"+p.name))
  probe=check(["go","test","-count=1","-timeout=30s","./..."],reconstructed)
  probe["kind"]="withheld-contract-probes"; checks.append(probe)
(arc/"verification.json").write_text(json.dumps(checks,indent=2)+"\n")
entry["source_sha256"]=hashes(work)
entry["reconstruction"]="byte-for-byte verified"
entry["artifact_sha256"]=hashes(arc)
entry["completed"]=True
manifest["stage"]="author outcomes archived; independent review/promotion pending"
(run/"manifest.json").write_text(json.dumps(manifest,indent=2)+"\n")
print(name,[(c.get("kind",c.get("command",[None])[-1]),c.get("exit_code",c.get("result"))) for c in checks])
