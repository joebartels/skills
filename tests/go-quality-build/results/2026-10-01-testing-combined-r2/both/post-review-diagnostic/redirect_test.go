package indexer

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
 if err:=os.WriteFile(path, []byte("prior snapshot\n"), 0600); err!=nil { t.Fatal(err) }
 tr:=new(diagnosticTransport); client:=&http.Client{Transport:tr}
 err:=Refresh(context.Background(), client, "http://diagnostic.invalid/start", path)
 got,readErr:=os.ReadFile(path); if readErr!=nil { t.Fatal(readErr) }
 if err==nil { t.Error("initial 302 must reject") }
 if tr.requests!=1 { t.Errorf("requests=%d, want 1", tr.requests) }
 if tr.closed!=1 { t.Errorf("closed bodies=%d, want 1", tr.closed) }
 if string(got)!="prior snapshot\n" { t.Errorf("prior bytes changed to %q", got) }
 if client.CheckRedirect!=nil { t.Error("borrowed client redirect configuration changed") }
}
