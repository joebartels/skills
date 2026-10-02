from pathlib import Path
import json, os, shutil, subprocess, tempfile, time, sys

run=Path.cwd()/'tests/go-quality-build/results'/sys.argv[1]
probe='''package indexer

import (
 "context"
 "io"
 "net/http"
 "os"
 "path/filepath"
 "strings"
 "testing"
)

type diagnosticTransport struct { requests, closed int }
type diagnosticBody struct { io.Reader; transport *diagnosticTransport }
func (b diagnosticBody) Close() error { b.transport.closed++; return nil }
func (tr *diagnosticTransport) RoundTrip(req *http.Request) (*http.Response, error) {
 tr.requests++
 status, body := http.StatusFound, "redirect body"
 header := make(http.Header)
 if req.URL.Path == "/start" { header.Set("Location", "/final") } else { status=http.StatusOK; body=`[{"key":"new","text":"value"}]` }
 return &http.Response{StatusCode:status, Header:header, Body:diagnosticBody{strings.NewReader(body), tr}, Request:req}, nil
}
func TestDiagnosticSingleRequestInitialStatus(t *testing.T) {
 path:=filepath.Join(t.TempDir(), "snapshot")
 if err:=os.WriteFile(path, []byte("prior snapshot\\n"), 0600); err!=nil { t.Fatal(err) }
 tr:=new(diagnosticTransport); client:=&http.Client{Transport:tr}
 err:=Refresh(context.Background(), client, "http://diagnostic.invalid/start", path)
 got,readErr:=os.ReadFile(path); if readErr!=nil { t.Fatal(readErr) }
 if err==nil { t.Error("initial 302 must reject") }
 if tr.requests!=1 { t.Errorf("requests=%d, want 1", tr.requests) }
 if tr.closed!=1 { t.Errorf("closed bodies=%d, want 1", tr.closed) }
 if string(got)!="prior snapshot\\n" { t.Errorf("prior bytes changed to %q", got) }
 if client.CheckRedirect!=nil { t.Error("borrowed client redirect configuration changed") }
}
'''
for case in json.loads((run/'manifest.json').read_text())['cases']:
 arc=run/case['id'];out=arc/'post-review-diagnostic';out.mkdir(exist_ok=True)
 (out/'redirect_test.go').write_text(probe)
 with tempfile.TemporaryDirectory(prefix='go-redirect-diagnostic-',dir='/private/tmp') as tmp:
  work=Path(tmp)/'module';shutil.copytree(arc/'candidate',work);shutil.copy2(out/'redirect_test.go',work/'diagnostic_redirect_test.go')
  cmd=['rtk','proxy','go','test','-run','^TestDiagnosticSingleRequestInitialStatus$','-count=1','-timeout=15s','.'];start=time.monotonic()
  r=subprocess.run(cmd,cwd=work,env=dict(os.environ,GOCACHE='/private/tmp/go-quality-testing-cache',GOTOOLCHAIN='local'),capture_output=True,text=True,timeout=25)
  record=dict(command=cmd,cwd=str(work),env={'GOCACHE':'/private/tmp/go-quality-testing-cache','GOTOOLCHAIN':'local'},stdout=r.stdout,stderr=r.stderr,exit_code=r.returncode,duration=time.monotonic()-start,scope='identical post-review contract probe through concrete http.Client policy and controlled RoundTripper; no socket/network claim; not one of the six frozen mutations')
  (out/'verification.json').write_text(json.dumps(record,indent=2)+'\n')
  print(case['id'],r.returncode,r.stdout.strip())
