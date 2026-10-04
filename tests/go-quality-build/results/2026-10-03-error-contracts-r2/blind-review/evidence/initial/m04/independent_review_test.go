package accountlookup_test

import (
 "errors"
 "fmt"
 "strconv"
 "strings"
 "testing"
 "example.com/accountlookup"
)

type reviewBackend func(string)(string,error)
func (f reviewBackend) Find(k string)(string,error){return f(k)}
type reviewPrivateError struct{ calls *int }
func (e *reviewPrivateError) Error() string { *e.calls++; return "private credential and SQL" }
func TestIndependentReviewLookupBoundary(t *testing.T){
 calls:=0; raw:=&reviewPrivateError{&calls}; key:=" x\n\x00\" "
 var gotKey string
 wrapped:=fmt.Errorf("internal: %w",raw)
 svc:=accountlookup.Service{Backend:reviewBackend(func(k string)(string,error){gotKey=k;return "unusable",wrapped})}
 calls=0
 got,err:=svc.Lookup(key)
 if got!=""||gotKey!=key||!errors.Is(err,accountlookup.ErrUnavailable){t.Fatalf("result=%q err=%v key=%q",got,err,gotKey)}
 if errors.Is(err,raw){t.Fatal("private identity escaped")};var private *reviewPrivateError;if errors.As(err,&private){t.Fatal("private type escaped")}
 if !strings.Contains(err.Error(),strconv.Quote(key))||strings.Contains(err.Error(),"private")||strings.ContainsAny(err.Error(),"\n\x00"){t.Fatalf("unsafe or absent context: %q",err.Error())}
 if calls!=0{t.Fatalf("backend Error evaluated while translating: %d",calls)}
 for _,bad:=range []string{""," \t\r\n", "\u2003\u00a0"}{
  call:=false;got,err:=(accountlookup.Service{Backend:reviewBackend(func(string)(string,error){call=true;return "",nil})}).Lookup(bad)
  if got!=""||err!=accountlookup.ErrInvalidKey||call{t.Fatalf("blank=%q got=%q err=%v called=%v",bad,got,err,call)}
 }
 for _,backendErr:=range []error{accountlookup.BackendMissing,fmt.Errorf("private: %w",accountlookup.BackendMissing),errors.Join(raw,accountlookup.BackendMissing)}{
  got,err:=(accountlookup.Service{Backend:reviewBackend(func(string)(string,error){return "unusable",backendErr})}).Lookup("x")
  if got!=""||err!=accountlookup.ErrMissing{t.Fatalf("missing identity changed: %q %v",got,err)}
 }
 got,err=(accountlookup.Service{Backend:reviewBackend(func(string)(string,error){return "",nil})}).Lookup("x")
 if got!=""||err!=nil{t.Fatalf("empty successful value rejected: %q %v",got,err)}
}
