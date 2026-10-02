from pathlib import Path
import json,subprocess
root=Path.cwd(); suite=root/'tests/go-quality-build/testing-combined/evals'; case=suite/'files/indexer-evolution';case.mkdir(parents=True)
files={
'go.mod':'module example.com/indexer\n\ngo 1.22\n',
'README.md':'''# Indexer evolution

This standard-library Go1.22 module is a small file-backed index library plus an existing command. Preserve exported signatures and existing put behavior; implement the new operations and high-quality regression tests. Keep the solution proportionate. No prescribed package, interface, testing framework or fixture layout.

Existing Open(root) returns an independent Index. Put(key,text) accepts lowercase ASCII alphabetic keys and arbitrary text, atomically replaces that key's file, and preserves the old file on failure. Get returns exact stored text. Independent roots may use identical keys.

New contracts:

- (*Index).ApplyCSV(io.Reader) error accepts CSV rows with exactly two fields, key and text, without a header. Keys obey Put's rule; text must contain non-whitespace. Process rows in order; duplicate keys replace. A parse, validation or write failure stops processing and keeps all accepted earlier rows, preserves a previously accepted value when its replacement is rejected, and does not write later rows. Validation errors retain ErrInvalidRecord identity. Do not invent all-or-nothing semantics.
- Record has lowercase JSON key/text fields. Refresh(ctx,client,endpoint,path) makes one GET using the supplied context/client. Only status 200 with exactly one JSON array of Records is accepted; trailing non-whitespace data rejects. Every key follows the key rule and every text is nonblank; retain order and original text. Publish compact JSON plus newline atomically at path. Empty array publishes [] plus newline. Fetch, status, read, decode, validation or publication failure keeps the exact prior destination bytes. Close an acquired response body on success and rejection. Validation retains ErrInvalidRecord identity. On the evaluated POSIX filesystem, a reader opened before replacement retains the complete old snapshot; a later opener sees the complete new snapshot. Windows replacement semantics are outside this task's execution claim.
- Serve(ctx,interval,refresh,release) owns a refresh lifecycle. interval must be positive, otherwise return ErrInvalidInterval before any callback or release. For a valid interval, honor already-canceled context without starting refresh. Otherwise run immediately and sequentially; after a successful callback, wait the interval measured from its completion before the next. Pass the caller's context to every callback. On cancellation or callback failure, join all started work before calling release exactly once and returning. Preserve callback and release error identities, including both together; ordinary parent cancellation alone returns nil. The callback may perform cooperative cancellation cleanup before returning; release must not overtake that cleanup. Invocations have independent state.
- Preserve `indexer put DIR KEY TEXT`: success exits 0 without stdout, failure/usage exits 2 with stderr. Add `indexer apply --dir DIR` reading CSV from stdin with the same status/stream convention and accepted-prefix effects. Add `indexer watch --url URL --file PATH --interval DURATION`: URL must be HTTP/HTTPS with a host, path nonempty, interval positive. It invokes Serve with Refresh and handles interrupt/termination by canceling and joining; close client idle connections only after refresh stops. Invalid startup arguments or failed first refresh exit 2 with a diagnostic and no success stdout. Successful refreshes recur until cancellation. Command helpers alone do not change these process contracts.

Verification should cover consequential success, rejection, retained state, actual command behavior and lifecycle paths. State material unsupported boundaries. Use finite test/process deadlines. Do not raise the Go minimum or install tools/dependencies.
''',
'index.go':'''package indexer

import (
 "errors"
 "io"
 "os"
 "path/filepath"
 "regexp"
)

var ErrInvalidRecord = errors.New("invalid record")
var ErrNotImplemented = errors.New("operation not implemented")
var keyPattern = regexp.MustCompile(`^[a-z]+$`)

type Index struct { root string }
func Open(root string) *Index { return &Index{root: root} }
func (i *Index) Put(key, text string) error {
 if !keyPattern.MatchString(key) { return ErrInvalidRecord }
 if err := os.MkdirAll(i.root, 0700); err != nil { return err }
 f, err := os.CreateTemp(i.root, ".record-*"); if err != nil { return err }
 defer os.Remove(f.Name())
 if _, err = f.WriteString(text); err != nil { f.Close(); return err }
 if err = f.Close(); err != nil { return err }
 return os.Rename(f.Name(), filepath.Join(i.root, key))
}
func (i *Index) Get(key string) (string, error) {
 if !keyPattern.MatchString(key) { return "", ErrInvalidRecord }
 b, err := os.ReadFile(filepath.Join(i.root,key)); return string(b), err
}
func (i *Index) ApplyCSV(r io.Reader) error { return ErrNotImplemented }
''',
'refresh.go':'''package indexer

import (
 "context"
 "net/http"
)

type Record struct { Key string `json:"key"`; Text string `json:"text"` }
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error { return ErrNotImplemented }
''',
'serve.go':'''package indexer

import (
 "context"
 "errors"
 "time"
)
var ErrInvalidInterval = errors.New("invalid interval")
func Serve(ctx context.Context, interval time.Duration, refresh func(context.Context) error, release func() error) error { return ErrNotImplemented }
''',
'index_test.go':'''package indexer_test

import (
 "errors"
 "testing"
 "example.com/indexer"
)
func TestExistingPutAndReplace(t *testing.T) {
 i := indexer.Open(t.TempDir())
 for _, text := range []string{"first", "x", ""} {
  if err := i.Put("alpha",text); err != nil { t.Fatal(err) }
  got, err := i.Get("alpha"); if err != nil || got != text { t.Fatalf("Get = %q, %v; want %q",got,err,text) }
 }
 if err := i.Put("../escape","bad"); !errors.Is(err,indexer.ErrInvalidRecord) { t.Fatalf("invalid key = %v",err) }
}
''',
'cmd/indexer/main.go':'''package main

import (
 "fmt"
 "os"
 "example.com/indexer"
)
func run(args []string) error {
 if len(args) != 4 || args[0] != "put" { return fmt.Errorf("usage: indexer put DIR KEY TEXT") }
 return indexer.Open(args[1]).Put(args[2],args[3])
}
func main() {
 if err := run(os.Args[1:]); err != nil { fmt.Fprintln(os.Stderr,err); os.Exit(2) }
}
''',
'cmd/indexer/main_test.go':'''package main

import "testing"
func TestExistingPut(t *testing.T) {
 if err := run([]string{"put", t.TempDir(), "alpha", "one"}); err != nil { t.Fatal(err) }
 if err := run(nil); err == nil { t.Fatal("missing usage rejection") }
}
'''
}
for name,text in files.items():p=case/name;p.parent.mkdir(parents=True,exist_ok=True);p.write_text(text)
subprocess.run(['rtk','proxy','gofmt','-w','.'],cwd=case,check=True)
suite.joinpath('evals.json').write_text(json.dumps(dict(skill_name='testing-combined',evals=[dict(id='indexer-evolution',prompt='Implement the indexer evolution and high-quality Go regression tests according to README.md. Preserve existing consumer/process contracts and Go1.22 support; keep the design proportionate.',expected_output='Ordered ingress, complete retained snapshots and host-owned refresh lifecycle with effective reliable contract tests.',assertions=['Actual command failure status and accepted-prefix effects are verified.','Complete publication, validation, caller cancellation, successful recurrence and joined release are implemented and meaningfully tested.','Neighboring writing guidance remains consistent without unnecessary test/production machinery.'],files=['files/indexer-evolution/'+n for n in sorted(files)] )]),indent=2)+'\n')
probe=suite/'controller-probes/indexer-evolution';probe.mkdir(parents=True)
probe.joinpath('contract_test.go').write_text('''package indexer_test

import (
 "context"
 "errors"
 "io"
 "net/http"
 "net/http/httptest"
 "os"
 "os/exec"
 "path/filepath"
 "strings"
 "testing"
 "time"
 "example.com/indexer"
)
func TestControllerAcceptedPrefix(t *testing.T) {
 root:=t.TempDir();i:=indexer.Open(root)
 if err:=i.ApplyCSV(strings.NewReader("a,old\\na,new\\nb,kept\\na,  \\nc,later\\n")); !errors.Is(err,indexer.ErrInvalidRecord) {t.Fatalf("invalid row: %v",err)}
 for key,want:=range map[string]string{"a":"new","b":"kept"} {got,err:=i.Get(key);if err!=nil||got!=want {t.Fatalf("%s=%q,%v want %q",key,got,err,want)}}
 if _,err:=i.Get("c");!errors.Is(err,os.ErrNotExist) {t.Fatalf("later row exists: %v",err)}
}
type controllerTransport func(*http.Request)(*http.Response,error)
func(f controllerTransport)RoundTrip(r *http.Request)(*http.Response,error){return f(r)}
func TestControllerSnapshotPublication(t *testing.T) {
 path:=filepath.Join(t.TempDir(),"index.json");old:=`[{"key":"old","text":"Previous"}]`+"\\n";if err:=os.WriteFile(path,[]byte(old),0600);err!=nil {t.Fatal(err)}
 opened,err:=os.Open(path);if err!=nil {t.Fatal(err)};defer opened.Close()
 body:=`[{"key":"a","text":"Alpha"},{"key":"b","text":"Beta"}]`
 client:=&http.Client{Transport:controllerTransport(func(r *http.Request)(*http.Response,error){if r.Method!="GET"{t.Errorf("method %s",r.Method)};return &http.Response{StatusCode:200,Body:io.NopCloser(strings.NewReader(body)),Header:make(http.Header)},nil})}
 if err:=indexer.Refresh(context.Background(),client,"http://example.invalid/index",path);err!=nil {t.Fatal(err)}
 got,err:=os.ReadFile(path);if err!=nil||string(got)!=body+"\\n" {t.Fatalf("new snapshot=%q,%v",got,err)}
 prior,err:=io.ReadAll(opened);if err!=nil||string(prior)!=old {t.Fatalf("opened prior snapshot=%q,%v",prior,err)}
 body=`[{"key":"a","text":"  "}]`;if err:=indexer.Refresh(context.Background(),client,"http://example.invalid/index",path);!errors.Is(err,indexer.ErrInvalidRecord){t.Fatalf("invalid text=%v",err)}
 retained,err:=os.ReadFile(path);if err!=nil||string(retained)!=string(got){t.Fatalf("rejection changed snapshot=%q,%v",retained,err)}
 body=`[{"key":"a","text":"Alpha"}] true`;if err:=indexer.Refresh(context.Background(),client,"http://example.invalid/index",path);err==nil {t.Fatal("trailing JSON accepted")}
}
func TestControllerHostJoinAndErrors(t *testing.T) {
 ctx,cancel:=context.WithCancel(context.Background());defer cancel();started:=make(chan struct{});cleaning:=make(chan struct{});allow:=make(chan struct{});finished:=make(chan struct{});released:=make(chan bool,1);result:=make(chan error,1);workErr:=errors.New("work");releaseErr:=errors.New("release")
 t.Cleanup(func(){cancel();select{case <-allow:default:close(allow)};select{case <-result:case <-time.After(5*time.Second):t.Error("host did not join")}})
 go func(){result<-indexer.Serve(ctx,20*time.Millisecond,func(c context.Context)error{close(started);<-c.Done();close(cleaning);<-allow;close(finished);return workErr},func()error{select{case <-finished:released<-true;default:released<-false};return releaseErr})}()
 select{case <-started:case <-time.After(3*time.Second):t.Fatal("callback not started")};cancel()
 select{case <-cleaning:case <-time.After(3*time.Second):t.Fatal("caller cancellation absent")}
 select{case <-released:t.Fatal("release before callback cleanup");case <-time.After(20*time.Millisecond):}
 close(allow);select{case err:=<-result:if !errors.Is(err,workErr)||!errors.Is(err,releaseErr){t.Fatalf("error identities=%v",err)};result<-nil;case <-time.After(3*time.Second):t.Fatal("host completion absent")}
 if !<-released {t.Fatal("release preceded completion")}
}
func TestControllerSuccessfulRecurrence(t *testing.T) {
 ctx,cancel:=context.WithCancel(context.Background());defer cancel();entered:=make(chan int,3);gate:=make(chan struct{});completed:=make(chan time.Time,1);second:=make(chan time.Time,1);result:=make(chan error,1);calls:=0
 t.Cleanup(func(){cancel();select{case <-gate:default:close(gate)};select{case <-result:case <-time.After(5*time.Second):t.Error("recurrence host did not join")}})
 go func(){result<-indexer.Serve(ctx,50*time.Millisecond,func(context.Context)error{calls++;entered<-calls;if calls==1{<-gate;completed<-time.Now()}else{second<-time.Now();cancel()};return nil},func()error{return nil})}()
 select{case <-entered:case <-time.After(3*time.Second):t.Fatal("no initial refresh")};time.Sleep(70*time.Millisecond);close(gate)
 finish:=<-completed;select{case next:=<-second:if gap:=next.Sub(finish);gap<45*time.Millisecond {t.Fatalf("recurrence gap %v < completion-relative interval",gap)};case <-time.After(3*time.Second):t.Fatal("no successful recurrence")}
}
func TestControllerActualProcessFailures(t *testing.T) {
 binary:=filepath.Join(t.TempDir(),"indexer");build:=exec.Command("go","build","-o",binary,"./cmd/indexer");if out,err:=build.CombinedOutput();err!=nil{t.Fatalf("build: %v %s",err,out)}
 dir:=t.TempDir();ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second);defer cancel();cmd:=exec.CommandContext(ctx,binary,"apply","--dir",dir);cmd.Stdin=strings.NewReader("a,first\\nb,  \\nc,later\\n");var stdout,stderr strings.Builder;cmd.Stdout=&stdout;cmd.Stderr=&stderr
 err:=cmd.Run();var exit *exec.ExitError;if !errors.As(err,&exit)||exit.ExitCode()!=2||stdout.Len()!=0||stderr.Len()==0{t.Fatalf("apply status/streams=%v/%q/%q",err,stdout.String(),stderr.String())};got,err:=os.ReadFile(filepath.Join(dir,"a"));if err!=nil||string(got)!="first"{t.Fatalf("process prefix=%q,%v",got,err)}
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.WriteHeader(503)}));defer server.Close();ctx2,cancel2:=context.WithTimeout(context.Background(),5*time.Second);defer cancel2();cmd=exec.CommandContext(ctx2,binary,"watch","--url",server.URL,"--file",filepath.Join(t.TempDir(),"index.json"),"--interval","10ms");stdout.Reset();stderr.Reset();cmd.Stdout=&stdout;cmd.Stderr=&stderr;err=cmd.Run();if !errors.As(err,&exit)||exit.ExitCode()!=2||stdout.Len()!=0||stderr.Len()==0{t.Fatalf("watch failed-start status/streams=%v/%q/%q",err,stdout.String(),stderr.String())}
}
''')
subprocess.run(['rtk','proxy','gofmt','-w',str(probe/'contract_test.go')],check=True)
(suite/'controller-probes/mutations.md').write_text('''# Frozen integrated mutation meanings

Candidate tests only; probes are never injected into mutation runs. Patch each candidate semantically at its own implementation, verify clean and compiled states, then interpret explicit contract failures. A compiler error, unrelated panic or overall timeout is not a detected regression.

- process-success-on-error: main retains diagnostic but exits 0 after apply/watch failure.
- host-release-before-join: cancellation returns/releases while an already-started refresh is still doing cooperative cleanup.
- direct-snapshot-truncation: replace atomic snapshot publication with in-place destination writes; an already-open old reader sees modified bytes.
- discarded-caller-context: refresh receives background context instead of the caller context.
- start-relative-recurrence: delay is measured from callback start so a long successful cycle has no required post-completion interval.
- omitted-text-validation: a nonempty key with blank text is accepted; useful separate field/state signal.

Targeted patches are controller diagnostics, not author input or a prescribed test form. POSIX snapshot observation is task-scoped; Windows runtime is unverified.
''')
print('Created fresh combined input',len(files),'files and withheld contract probes')
