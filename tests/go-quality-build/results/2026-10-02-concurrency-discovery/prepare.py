"""Fixed concurrency fixture/controller preparation; no authors or runtime changes."""
from pathlib import Path
import hashlib,json,os,re,shutil,subprocess,tempfile,time
ROOT=Path(__file__).resolve().parents[4];OUT=Path(__file__).resolve().parent;PREP=OUT/"preparation";PREP.mkdir(exist_ok=True)
EVALS=ROOT/"tests/go-quality-build/go-concurrency-and-ownership/evals"
MIN="/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go"
ENV={"GOCACHE":"/private/tmp/go-quality-concurrency-cache","GOTOOLCHAIN":"local"}
cases=[x["id"] for x in json.loads((EVALS/"evals.json").read_text())["evals"]]
def run(argv,cwd,extra=None):
 env=os.environ.copy();env.update(ENV);env.update(extra or {});start=time.monotonic();p=subprocess.run(["rtk","proxy"]+argv,cwd=cwd,env=env,capture_output=True,text=True,timeout=90)
 return {"argv":["rtk","proxy"]+argv,"cwd":str(cwd),"environment_overrides":ENV|dict(extra or {}),"status":p.returncode,"stdout":p.stdout,"stderr":p.stderr,"elapsed_seconds":round(time.monotonic()-start,3)}
def controller(case,work):
 destination=work/("cmd/pipeline" if case=="cli-pipeline-stop" else "")/"controller_contract_test.go";shutil.copyfile(EVALS/"controller-probes"/case/"contract_test.go",destination)
def checks(work,minimum=True):
 results=[run(a,work) for a in [["go","build","-o","/dev/null","./..."],["go","test","-count=1","-timeout=30s","./..."],["go","vet","./..."],["gofmt","-l","."]]]
 if minimum:results.append(run([MIN,"test","-count=1","-timeout=30s","./..."],work,{"GOQUALITY_GO":MIN}))
 return results
for directory in [EVALS/"files"/c for c in cases]+[EVALS/"controller-probes"/c for c in cases]:
 result=run(["gofmt","-w","."],directory);assert result["status"]==0,result
versions=[run(["go","version"],ROOT),run([MIN,"version"],ROOT)];(PREP/"toolchains.json").write_text(json.dumps(versions,indent=2)+"\n")
ordinary={};red={}
for case in cases:
 ordinary[case]=checks(EVALS/"files"/case);assert all(c["status"]==0 for c in ordinary[case]),ordinary[case]
 work=Path(tempfile.mkdtemp(prefix="go-requested-",dir="/private/tmp"));shutil.copytree(EVALS/"files"/case,work,dirs_exist_ok=True);controller(case,work);red[case]=run(["go","test","-count=1","-timeout=30s","./..."],work);assert red[case]["status"]!=0,case
 print(case,"original ordinary",[x["status"] for x in ordinary[case]],"requested",red[case]["status"],flush=True)
(PREP/"original-ordinary.json").write_text(json.dumps(ordinary,indent=2)+"\n");(PREP/"requested-behavior-red.json").write_text(json.dumps(red,indent=2)+"\n")
