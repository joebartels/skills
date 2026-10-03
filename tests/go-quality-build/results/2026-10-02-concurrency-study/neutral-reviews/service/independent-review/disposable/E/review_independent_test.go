package workers
import("context";"errors";"sync";"sync/atomic";"testing";"time")
type independentLease struct {run func(context.Context,Job) error; close func() error}
func(l independentLease)Run(c context.Context,j Job)error{return l.run(c,j)}
func(l independentLease)Close()error{return l.close()}
func independentWait[T any](t *testing.T,c <-chan T,what string)T{t.Helper();select{case v:=<-c:return v;case <-time.After(time.Second):t.Fatalf("missing event: %s",what);var z T;return z}}
func TestIndependentCloseFailureStopsNextAdmission(t *testing.T){
 jobs:=make(chan Job,3);jobs<-1;jobs<-2;jobs<-3;close(jobs);sentinel:=errors.New("Close failure");var opened,closed atomic.Int32
 err:=Serve(context.Background(),jobs,1,func(context.Context,Job)(Lease,error){opened.Add(1);return independentLease{run:func(context.Context,Job)error{return nil},close:func()error{closed.Add(1);return sentinel}},nil})
 if !errors.Is(err,sentinel)||opened.Load()!=1||closed.Load()!=1{t.Errorf("after observed Close failure err=%v opened=%d closed=%d; want retained Close cause and one acquisition/release",err,opened.Load(),closed.Load())}
}
func TestIndependentAcquiredAfterCancelSkipsRun(t *testing.T){
 ctx,cancel:=context.WithCancel(context.Background());defer cancel();jobs:=make(chan Job,1);jobs<-1;close(jobs);var runs,closes atomic.Int32
 err:=Serve(ctx,jobs,1,func(context.Context,Job)(Lease,error){cancel();return independentLease{run:func(context.Context,Job)error{runs.Add(1);return nil},close:func()error{closes.Add(1);return nil}},nil})
 if !errors.Is(err,context.Canceled)||runs.Load()!=0||closes.Load()!=1{t.Errorf("acquired after cancellation err=%v runs=%d closes=%d; want canceled, zero Runs, one Close",err,runs.Load(),closes.Load())}
}
func TestIndependentCallerCancellationDuringFailureCleanup(t *testing.T){
 ctx,cancel:=context.WithCancel(context.Background());defer cancel();jobs:=make(chan Job,1);jobs<-1;close(jobs);runErr,closeErr:=errors.New("Run failure"),errors.New("Close failure");entered,release:=make(chan struct{}),make(chan struct{});done:=make(chan error,1);var once sync.Once
 t.Cleanup(func(){cancel();once.Do(func(){close(release)})})
 go func(){done<-Serve(ctx,jobs,1,func(context.Context,Job)(Lease,error){return independentLease{run:func(context.Context,Job)error{return runErr},close:func()error{close(entered);<-release;return closeErr}},nil})}()
 independentWait(t,entered,"Close entered after Run failure");cancel();once.Do(func(){close(release)});err:=independentWait(t,done,"failure cleanup joined")
 for _,want:=range []error{runErr,closeErr,context.Canceled}{if !errors.Is(err,want){t.Errorf("missing %v after cancellation during held Close: %v",want,err)}}
}
func TestIndependentHealthyCohortNotCanceled(t *testing.T){
 jobs:=make(chan Job,2);jobs<-1;jobs<-2;close(jobs);started:=make(chan context.Context,2);release:=make(chan struct{});done:=make(chan error,1);var once sync.Once
 t.Cleanup(func(){once.Do(func(){close(release)})})
 go func(){done<-Serve(context.Background(),jobs,2,func(context.Context,Job)(Lease,error){return independentLease{run:func(c context.Context,_ Job)error{started<-c;<-release;return c.Err()},close:func()error{return nil}},nil})}()
 first:=independentWait(t,started,"first Run");independentWait(t,started,"second Run")
 select{case <-first.Done():t.Error("healthy full cohort was canceled before successful work was released");case <-time.After(100*time.Millisecond):}
 once.Do(func(){close(release)});if err:=independentWait(t,done,"healthy cohort completion");err!=nil{t.Errorf("uncanceled successful input returned failure: %v",err)}
}
func TestIndependentCompletedRunsRetainCapacity(t *testing.T){
 ctx,cancel:=context.WithCancel(context.Background());defer cancel();jobs:=make(chan Job);returned:=make(chan Job,8);done:=make(chan error,1);var alive,peak,closes atomic.Int32
 go func(){done<-Serve(ctx,jobs,2,func(context.Context,Job)(Lease,error){n:=alive.Add(1);for old:=peak.Load();n>old;old=peak.Load(){if peak.CompareAndSwap(old,n){break}};return independentLease{run:func(_ context.Context,j Job)error{returned<-j;return nil},close:func()error{alive.Add(-1);closes.Add(1);return nil}},nil})}()
 for j:=Job(1);j<=8;j++{select{case jobs<-j:case <-time.After(time.Second):cancel();close(jobs);independentWait(t,done,"paced capacity cleanup");t.Fatal("paced job admission blocked")};independentWait(t,returned,"paced Run returned");time.Sleep(20*time.Millisecond)}
 close(jobs);err:=independentWait(t,done,"paced finite input completed");if err!=nil||peak.Load()>2||alive.Load()!=0||closes.Load()!=8{t.Errorf("acquired leases until Close: err=%v peak=%d alive=%d closes=%d; want nil, peak<=2, zero alive, eight closes",err,peak.Load(),alive.Load(),closes.Load())}
}
func TestIndependentLaterAvailableInputStarts(t *testing.T){
 ctx,cancel:=context.WithCancel(context.Background());defer cancel();jobs:=make(chan Job);started:=make(chan Job,2);release:=make(chan struct{});done:=make(chan error,1);var once sync.Once
 t.Cleanup(func(){cancel();once.Do(func(){close(release)})})
 go func(){done<-Serve(ctx,jobs,2,func(context.Context,Job)(Lease,error){return independentLease{run:func(_ context.Context,j Job)error{started<-j;<-release;return nil},close:func()error{return nil}},nil})}()
 jobs<-1;independentWait(t,started,"first sparse Run");sent:=make(chan struct{});go func(){select{case jobs<-2:case <-ctx.Done():};close(sent)}()
 missing:=false;select{case <-started:case <-time.After(250*time.Millisecond):missing=true}
 once.Do(func(){close(release)});independentWait(t,sent,"second sender completed");close(jobs);independentWait(t,done,"later-input cleanup joined")
 if missing{t.Error("second available job did not start while first Run held only one of two capacity slots")}
}
