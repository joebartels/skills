package lineexport
import("bytes";"errors";"io";"strings";"testing")
type probeWriter struct{buf bytes.Buffer;writeErr,closeErr error;closed int}
func(w *probeWriter)Write(p []byte)(int,error){if w.writeErr!=nil{return 0,w.writeErr};return w.buf.Write(p)}
func(w *probeWriter)String()string{return w.buf.String()}
func(w *probeWriter)Len()int{return w.buf.Len()}
func(w *probeWriter)Close()error{w.closed++;return w.closeErr}
type probeFailReader struct{done bool;cause error}
func(r *probeFailReader)Read(p []byte)(int,error){if r.done{return 0,io.EOF};r.done=true;return copy(p,"a\n"),r.cause}
func TestContractCompletionIdentity(t *testing.T){primary:=errors.New("read failed");finish:=errors.New("close failed");for _,tc:=range []struct{name string;primary,finish error}{{"primary",primary,nil},{"finish",nil,finish},{"both",primary,finish},{"success",nil,nil}}{t.Run(tc.name,func(t *testing.T){var r io.Reader=strings.NewReader("a\n");if tc.primary!=nil{r=&probeFailReader{cause:tc.primary}};w:=&probeWriter{closeErr:tc.finish};n,err:=Export(r,w);if w.closed!=1||n!=1||w.String()!="a\n"{t.Fatalf("ownership/prefix: %d %d %q",w.closed,n,w.String())};switch{case tc.primary!=nil&&tc.finish!=nil:if !errors.Is(err,primary)||!errors.Is(err,finish){t.Fatal(err)};case tc.primary!=nil:if err!=primary{t.Fatalf("changed direct primary identity: %v",err)};case tc.finish!=nil:if err!=finish{t.Fatalf("changed direct close identity: %v",err)};default:if err!=nil{t.Fatal(err)}}})}}
func TestContractLimitAndNegative(t *testing.T){w:=new(probeWriter);n,err:=ExportLimit(strings.NewReader("a\nb\nc\n"),w,2);if n!=2||err!=nil||w.closed!=1||w.String()!="a\nb\n"{t.Fatalf("limit %d %v %q",n,err,w.String())};w=new(probeWriter);n,err=ExportLimit(strings.NewReader("a\n"),w,-1);if n!=0||err==nil||w.closed!=1||w.Len()!=0{t.Fatalf("negative %d %v %d",n,err,w.closed)}}
func TestContractWriteFailure(t *testing.T){cause:=errors.New("write failed");w:=&probeWriter{writeErr:cause};n,err:=Export(strings.NewReader("a\n"),w);if n!=0||err!=cause||w.closed!=1{t.Fatalf("%d %v %d",n,err,w.closed)}}
