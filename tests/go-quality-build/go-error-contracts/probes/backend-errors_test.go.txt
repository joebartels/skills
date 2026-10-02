package accountlookup
import("errors";"strings";"testing")
type probeBackend struct{value string;err error;calls int}
func(b *probeBackend)Find(string)(string,error){b.calls++;return b.value,b.err}
type secretError struct{}
func(*secretError)Error()string{return "password=secret SQL SELECT private"}
func TestContractDomainExposure(t *testing.T){raw:=&secretError{};b:=&probeBackend{value:"unusable",err:raw};got,err:=(Service{Backend:b}).Lookup("alice");var exposed *secretError;if got!=""||!errors.Is(err,ErrUnavailable)||errors.Is(err,raw)||errors.As(err,&exposed)||strings.Contains(err.Error(),"secret")||strings.Contains(err.Error(),"SELECT"){t.Fatalf("exposure: %q %v",got,err)};if !strings.Contains(err.Error(),"alice"){t.Fatal("missing operation key context")}}
func TestContractDirectSentinelAndValidation(t *testing.T){b:=&probeBackend{err:BackendMissing};_,err:=(Service{Backend:b}).Lookup("u");if err!=ErrMissing{t.Fatalf("direct identity changed: %v",err)};b.calls=0;_,err=(Service{Backend:b}).Lookup("  ");if !errors.Is(err,ErrInvalidKey)||b.calls!=0{t.Fatalf("validation %v calls=%d",err,b.calls)}}
