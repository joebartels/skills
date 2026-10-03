"""Per-run reconstruction/check capture for the fixed context study outcomes, not an author."""
from pathlib import Path
import concurrent.futures,hashlib,json,os,re,shutil,subprocess,sys,tempfile,time
ROOT=Path(__file__).resolve().parents[4];OUT=Path(__file__).resolve().parent;EVALS=ROOT/"tests/go-quality-build/go-context-and-deadlines/evals"
MIN="/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go"
manifest=json.loads((OUT/"manifest.json").read_text())
rows={a["number"]:a for a in manifest["attempts"]}
numbers=[int(n) for n in sys.argv[1:]];assert numbers and all(n in rows and rows[n]["status"]=="completed" for n in numbers)
ENV={"GOCACHE":"/private/tmp/go-quality-context-cache","GOTOOLCHAIN":"local","GOQUALITY_GO":MIN}
def run(argv,cwd,extra=None):
 env=os.environ.copy();env.update(ENV);env.update(extra or {});start=time.monotonic()
 p=subprocess.run(["rtk","proxy"]+argv,cwd=cwd,env=env,capture_output=True,text=True,timeout=90)
 return {"argv":["rtk","proxy"]+argv,"cwd":str(cwd),"environment_overrides":ENV|dict(extra or {}),"status":p.returncode,"stdout":p.stdout,"stderr":p.stderr,"elapsed_seconds":round(time.monotonic()-start,3)}
def mutation(n,case,name,changes,test,original_held_status):
 archive=OUT/f"attempt-{n:02}";work=Path(tempfile.mkdtemp(prefix="go-mutation-",dir="/private/tmp"));shutil.copytree(archive/"source",work,dirs_exist_ok=True)
 actual=[]
 for file,old,new in changes:
  p=work/file;s=p.read_text()
  if old not in s:return {"id":name,"not_applied":"Different implementation; concrete substitution unavailable","file":file,"required_old":old}
  p.write_text(s.replace(old,new,1));actual.append({"file":file,"old":old,"new":new})
 run(["gofmt","-w","."],work)
 result={"id":name,"substitutions":actual,"compile":run(["go","test","-run","^$","./..."],work),"author_tests":run(["go","test","-count=1","-timeout=30s","./..."],work),"original_held_named_status":original_held_status}
 shutil.copyfile(EVALS/"controller-probes"/case/"contract_test.go",work/"contract_controller_test.go")
 result["held_tests"]=run(["go","test","-run","^"+test+"$","-count=1","-timeout=30s","./..."],work)
 result["qualified"]=result["compile"]["status"]==0 and original_held_status==0 and result["held_tests"]["status"]!=0
 return result
def verify(n):
 row=rows[n];case=row["case"];archive=OUT/row["directory"];work=Path(tempfile.mkdtemp(prefix="go-check-",dir="/private/tmp"));shutil.copytree(EVALS/"files"/case,work,dirs_exist_ok=True)
 reconstruction=run(["git","apply",str(archive/"source.patch")],work);assert reconstruction["status"]==0,reconstruction
 hashes={str(p.relative_to(work)):hashlib.sha256(p.read_bytes()).hexdigest() for p in work.rglob("*") if p.is_file()};assert hashes==json.loads((archive/"source-hashes.json").read_text()),n
 ordinary=[reconstruction]
 for argv in [["go","version"],[MIN,"version"],["go","build","./..."],["go","test","-count=1","-timeout=30s","./..."],["go","vet","./..."],["gofmt","-l","."]]:ordinary.append(run(argv,work))
 extra={"CGO_ENABLED":"0"} if case=="service-total-budget" else {}
 ordinary.append(run([MIN,"test","-count=1","-timeout=30s","./..."],work,extra))
 shutil.copyfile(EVALS/"controller-probes"/case/"contract_test.go",work/"contract_controller_test.go")
 held=[run(["go","test","-count=1","-timeout=30s","./..."],work),run([MIN,"test","-count=1","-timeout=30s","./..."],work,extra)]
 if case!="not-context-work":held.append(run(["go","test","-race","-shuffle=on","-count=3","-timeout=60s","./..."],work))
 specifications=[]
 if case=="library-call-cause":
  text=(archive/"source/process.go").read_text();match=re.search(r"func Process\([\s\S]+?\{\n",text);assert match,n
  old=match.group(0);new=old+"\tmutOriginalApply := apply\n\tapply = func(mutCtx context.Context, mutJob int) error {\n\t\tmutErr := mutOriginalApply(mutCtx, mutJob)\n\t\tif mutErr != nil && context.Cause(mutCtx) != nil { _ = mutErr == context.Cause(mutCtx) }\n\t\treturn mutErr\n\t}\n"
  specifications.append(("unsafe-cause-equality",[("process.go",old,new)],"TestCauseReturnedByCallback"))
  changes=[]
  for file in (archive/"source").glob("*.go"):
   if file.name.endswith("_test.go"):continue
   text=file.read_text()
   for match in re.finditer(r"context\.Cause\([^()]*\)",text):changes.append((file.name,match.group(0),"error(nil)"))
  if changes:specifications.append(("lost-custom-cause",changes,"TestCustomCauseAndIndependentFailure"))
 elif case=="service-total-budget":
  old="for _, endpoint := range endpoints {\n";new=old+"\t\tvar mutCancel context.CancelFunc\n\t\ttotalCtx, mutCancel = context.WithTimeout(ctx, total)\n\t\tdefer mutCancel()\n"
  specifications.append(("lost-total-budget",[("stages.go",old,new)],"TestTotalAndEarlierParentDeadline"))
 elif case=="not-context-work":
  specifications.append(("control-upper-regression",[("range.go","value <= high","value < high")],"TestInclusiveUpperBound"))
 elif case=="cli-finalization-budget":
  text=(archive/"source/run.go").read_text();start=text.find("context.WithTimeout(")
  if start>=0:
   depth=0;end=start+len("context.WithTimeout");opening=end;comma=None
   for pos in range(opening,len(text)):
    if text[pos]=="(":depth+=1
    elif text[pos]==")":
     depth-=1
     if depth==0:end=pos+1;break
    elif text[pos]=="," and depth==1:comma=pos
   assert comma is not None
   parent=text[opening+1:comma];budget=text[comma+1:end-1].strip();old=text[start:end]
   new="func() (context.Context, context.CancelFunc) { _ = "+budget+"; return context.WithCancel("+parent+") }()"
   specifications.append(("unbounded-finalization",[("run.go",old,new)],"TestFinalizationDeadline"))
 mutations=[]
 for name,changes,test in specifications:
  original=run(["go","test","-run","^"+test+"$","-count=1","-timeout=30s","./..."],work)
  mutations.append(mutation(n,case,name,changes,test,original["status"]))
 supplementary=[]
 if case=="service-total-budget":
  source=work/"supplementary_contract_test.go"
  source.write_text('package stages\nimport("context";"errors";"net/http";"testing";"time")\nfunc TestSupplementaryCanceledEmpty(t *testing.T){ctx,cancel:=context.WithCancel(context.Background());cancel();got,err:=FetchAll(ctx,&http.Client{},nil,time.Second,time.Second);if err!=nil||len(got)!=0{t.Fatalf("empty endpoints must succeed without calls: bodies=%v err=%v",got,err)}}\nfunc TestSupplementaryIndependentCustomCause(t *testing.T){ctx,cancel:=context.WithCancelCause(context.Background());cause:=errors.New("custom stop");failure:=errors.New("independent transport failure");client:=&http.Client{Transport:transportFunc(func(r *http.Request)(*http.Response,error){cancel(cause);return nil,failure})};_,err:=FetchAll(ctx,client,[]string{"http://example.test/a"},time.Second,time.Second);if !errors.Is(err,failure)||!errors.Is(err,cause)||!errors.Is(err,context.Canceled){t.Fatalf("missing independent/cause/classification: %v",err)}}\n')
  run(["gofmt","-w",source.name],work)
  (archive/"supplementary-contract-probes.go.txt").write_bytes(source.read_bytes())
  supplementary=[run(["go","test","-run","^TestSupplementary","-count=1","-timeout=30s","./..."],work)]
 result={"reconstruction_hashes_match":True,"ordinary_checks":ordinary,"held_contract_checks":held,"frozen_mutation_sensitivity":mutations,"supplementary_after_exposure":supplementary,"supplementary_limit":"Explicit README empty-success and failure classification combinations; introduced after exposure, not frozen mutations or new benefit primary."}
 (archive/"verification.json").write_text(json.dumps(result,indent=2)+"\n")
 print(n,case,"ordinary",[x["status"] for x in ordinary],"held",[x["status"] for x in held],"mutation",[(x["id"],x.get("compile",{}).get("status"),x.get("author_tests",{}).get("status"),x.get("qualified")) for x in mutations],"supplementary",[x["status"] for x in supplementary],flush=True)
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:list(pool.map(verify,numbers))
