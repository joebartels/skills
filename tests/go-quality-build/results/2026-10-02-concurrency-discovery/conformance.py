"""Controller-only solutions and four fixed mutation qualification checks; excluded from benefit."""
from pathlib import Path
import hashlib,json,os,re,shutil,subprocess,tempfile,time
ROOT=Path(__file__).resolve().parents[4];OUT=Path(__file__).resolve().parent;PREP=OUT/"preparation";EVALS=ROOT/"tests/go-quality-build/go-concurrency-and-ownership/evals"
MIN="/private/tmp/go-quality-toolchains/golang.org/toolchain@v0.0.1-go1.22.12.darwin-arm64/bin/go"
ENV={"GOCACHE":"/private/tmp/go-quality-concurrency-cache","GOTOOLCHAIN":"local"}
def run(argv,cwd,extra=None):
 env=os.environ.copy();env.update(ENV);env.update(extra or {});start=time.monotonic();p=subprocess.run(["rtk","proxy"]+argv,cwd=cwd,env=env,capture_output=True,text=True,timeout=90)
 return {"argv":["rtk","proxy"]+argv,"cwd":str(cwd),"environment_overrides":ENV|dict(extra or {}),"status":p.returncode,"stdout":p.stdout,"stderr":p.stderr,"elapsed_seconds":round(time.monotonic()-start,3)}
SOLUTIONS={
"library-shared-state":("totals.go",'''package totals
import "sync"
type Snapshot struct{Count int64;Sum int64}
type Totals struct{mu sync.Mutex;count,sum int64}
func(t *Totals)Add(delta int64){t.mu.Lock();defer t.mu.Unlock();t.count++;t.sum+=delta}
func(t *Totals)Snapshot()Snapshot{t.mu.Lock();defer t.mu.Unlock();return Snapshot{t.count,t.sum}}
'''),
"service-owned-workers":("host.go",'''package workers
import("context";"errors";"sync")
type Job int
type Lease interface{Run(context.Context,Job)error;Close()error}
func Serve(ctx context.Context,jobs <-chan Job,limit int,open func(context.Context,Job)(Lease,error))error{
 if err:=ctx.Err();err!=nil{return err}
 child,cancel:=context.WithCancel(ctx);defer cancel()
 var runs sync.WaitGroup;var mu sync.Mutex;var failures []error
 collect:=func(err error){if err==nil{return};mu.Lock();failures=append(failures,err);mu.Unlock();cancel()}
 var cohort []Lease
 finish:=func(){runs.Wait();for _,lease:=range cohort{collect(lease.Close())};cohort=nil}
 stopping:=false
 admission:for{
  if len(cohort)==limit{finish();if child.Err()!=nil{stopping=true;break}}
  var job Job;var ok bool
  select{case <-child.Done():stopping=true;break admission;case job,ok=<-jobs:}
  if !ok{break};if child.Err()!=nil{stopping=true;break}
  lease,err:=open(child,job);if err!=nil{collect(err);break};cohort=append(cohort,lease)
  if child.Err()!=nil{stopping=true;break}
  runs.Add(1);go func(l Lease,j Job){defer runs.Done();collect(l.Run(child,j))}(lease,job)
 }
 finish();if stopping{collect(ctx.Err())};mu.Lock();defer mu.Unlock();return errors.Join(failures...)
}
'''),
"not-concurrency-work":("sum.go",'''package positive
func sumPositive(values []int)int{sum:=0;for _,value:=range values{if value>0{sum+=value}};return sum}
'''),
"cli-pipeline-stop":("cmd/pipeline/pipeline.go",'''package main
import("context";"errors")
type pipelineResult struct{err error;stopped bool}
func runPipeline(ctx context.Context,capacity int,produce func(context.Context,chan<- int)error,consume func(context.Context,<-chan int)error)error{
 child,cancel:=context.WithCancel(ctx);defer cancel();output:=make(chan int,capacity)
 producerDone:=make(chan pipelineResult,1);consumerDone:=make(chan pipelineResult,1)
 go func(){producerErr:=produce(child,output);stopped:=child.Err()!=nil;close(output);if producerErr!=nil{cancel()};producerDone<-pipelineResult{producerErr,stopped}}()
 go func(){consumerErr:=consume(child,output);stopped:=child.Err()!=nil;cancel();consumerDone<-pipelineResult{consumerErr,stopped}}()
 producer,consumer:=<-producerDone,<-consumerDone
 normalize:=func(result pipelineResult)error{if result.stopped&&(result.err==context.Canceled||result.err==context.DeadlineExceeded){return nil};return result.err}
 return errors.Join(normalize(producer),normalize(consumer),ctx.Err())
}
''')}
GREEN={}
for case,(name,source) in SOLUTIONS.items():
 work=PREP/"conformance"/case
 if work.exists():shutil.rmtree(work)
 shutil.copytree(EVALS/"files"/case,work);(work/name).write_text(source)
 destination=work/("cmd/pipeline" if case=="cli-pipeline-stop" else "")/"controller_contract_test.go";shutil.copyfile(EVALS/"controller-probes"/case/"contract_test.go",destination)
 assert run(["gofmt","-w","."],work)["status"]==0
 checks=[run(a,work,{"GOQUALITY_GO":MIN} if a[0]==MIN else None) for a in [["go","build","./..."],["go","test","-count=1","-timeout=30s","./..."],["go","vet","./..."],["gofmt","-l","."],[MIN,"test","-count=1","-timeout=30s","./..."]]]
 if case!="not-concurrency-work":checks.append(run(["go","test","-race","-shuffle=on","-count=3","-timeout=60s","./..."],work))
 GREEN[case]=checks;print(case,"controller",[c["status"] for c in checks],flush=True)
(PREP/"conformance-green.json").write_text(json.dumps(GREEN,indent=2)+"\n")
assert all(c["status"]==0 and (c["argv"][2]!="gofmt" or not c["stdout"].strip()) for checks in GREEN.values() for c in checks),"Controller feasibility failed; preserve evidence before repair."
MUTATIONS=[]
for case,meaning,target in [("library-shared-state","split-compound-publication","TestControllerCoherentSnapshots"),("service-owned-workers","release-before-completion","TestControllerPartialStartAndBlockedAdmission"),("service-owned-workers","unresponsive-admission","TestControllerCancellationWithOpenInput"),("cli-pipeline-stop","unjoined-early-completion","TestControllerEarlyConsumerAndProducerFailure")]:
 original=PREP/"conformance"/case;work=Path(tempfile.mkdtemp(prefix="go-concurrency-mutation-",dir="/private/tmp"));shutil.copytree(original,work,dirs_exist_ok=True)
 file=work/SOLUTIONS[case][0];old=file.read_text()
 if meaning=="split-compound-publication":
  new='''package totals
import("runtime";"sync/atomic")
type Snapshot struct{Count int64;Sum int64}
type Totals struct{count,sum atomic.Int64}
func(t *Totals)Add(delta int64){t.count.Add(1);runtime.Gosched();t.sum.Add(delta)}
func(t *Totals)Snapshot()Snapshot{count:=t.count.Load();runtime.Gosched();return Snapshot{count,t.sum.Load()}}
'''
 elif meaning=="release-before-completion":new=old.replace("runs.Wait()","/* skipped completion join */",1)
 elif meaning=="unresponsive-admission":
  new,n=re.subn(r"select \{\s*case <-child.Done\(\):\s*stopping = true\s*break admission\s*case job, ok = <-jobs:\s*\}","job, ok = <-jobs",old,count=1);assert n==1
  # The admission label no longer has a labeled break after this substitution.
  new=new.replace("admission:\n", "",1)
 else:
  needle="consumerErr := consume(child, output)\n\t\tstopped := child.Err() != nil\n\t\tcancel()";assert needle in old
  new=old.replace(needle,"consumerErr := consume(child, output)\n\t\tstopped := child.Err() != nil\n\t\tif consumerErr != nil { cancel() }",1)
 assert old!=new;file.write_text(new);assert run(["gofmt","-w","."],work)["status"]==0
 compile=run(["go","test","-run","^$","./..."],work);probe=run(["go","test","-run","^"+target+"$","-count=1","-timeout=30s","./..."],work)
 row={"case":case,"meaning":meaning,"target":target,"file":str(file.relative_to(work)),"old":old,"new":file.read_text(),"compile":compile,"held_named":probe,"qualified":compile["status"]==0 and probe["status"]!=0};MUTATIONS.append(row);print(meaning,"compile",compile["status"],"held",probe["status"],flush=True)
(PREP/"mutation-feasibility.json").write_text(json.dumps(MUTATIONS,indent=2)+"\n");assert len(MUTATIONS)==4 and all(m["qualified"] for m in MUTATIONS)
