import subprocess,json,time,os
from pathlib import Path
root=Path(__file__).resolve().parent
minimum="/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go"
env={**os.environ,"GOCACHE":str(root/"cache"),"GOTOOLCHAIN":"local","GOWORK":"off","GOFLAGS":"-mod=readonly"}
results=[]
def run(c,name,args):
 cwd=root/"disposable"/c;argv=["rtk","proxy"]+args;start=time.monotonic()
 try:
  p=subprocess.run(argv,cwd=cwd,env=env,text=True,capture_output=True,timeout=70); d={"candidate":c,"name":name,"argv":argv,"cwd":str(cwd),"environment_overrides":{k:env[k] for k in ["GOCACHE","GOTOOLCHAIN","GOWORK","GOFLAGS"]},"status":p.returncode,"stdout":p.stdout,"stderr":p.stderr,"elapsed_seconds":round(time.monotonic()-start,3)}
 except subprocess.TimeoutExpired as e:
  d={"candidate":c,"name":name,"argv":argv,"cwd":str(cwd),"status":"external timeout","stdout":str(e.stdout),"stderr":str(e.stderr),"elapsed_seconds":round(time.monotonic()-start,3)}
 results.append(d);(root/"evidence"/"commands.json").write_text(json.dumps(results,indent=2)+"\n")
 (root/"evidence"/(c+"-"+name+".log")).write_text(json.dumps({k:v for k,v in d.items() if k not in ["stdout","stderr"]},indent=2)+"\nSTDOUT\n"+d["stdout"]+"\nSTDERR\n"+d["stderr"])
 print(c,name,d["status"],d["stdout"].strip()[:600],flush=True)
for c in "ABCDEF":
 for label,go in [("host","go"),("minimum",minimum)]:
  run(c,label+"-version",[go,"version"])
  run(c,label+"-authored",[go,"test","-run","^TestServe","-count=1","-timeout=20s","./..."])
  run(c,label+"-independent",[go,"test","-run","^TestIndependent","-count=1","-timeout=20s","./..."])
  run(c,label+"-contract",[go,"test","-run","^TestController|^TestReview","-count=1","-timeout=20s","./..."])
 run(c,"host-vet",["go","vet","./..."])
 run(c,"host-build",["go","build","-o","/dev/null","./..."])
 run(c,"format",["gofmt","-l","host.go","host_test.go"])
 run(c,"host-race",["go","test","-race","-shuffle=on","-run","^TestServe|^TestIndependent|^TestController|^TestReview","-count=3","-timeout=60s","./..."])
 run(c,"minimum-graph",[minimum,"list","-m","all"])
