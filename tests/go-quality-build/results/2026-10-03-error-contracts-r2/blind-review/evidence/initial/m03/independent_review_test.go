package lineexport_test
import("bytes";"errors";"io";"testing";"example.com/lineexport")
type reviewExportReader struct{parts []string;reads int;final error;written *int;checkOrder bool;closes int}
func(r *reviewExportReader)Read(p []byte)(int,error){
 if r.checkOrder&&r.reads>*r.written{return 0,errors.New("read before prior record write")}
 r.reads++; if len(r.parts)==0{return 0,r.final};n:=copy(p,r.parts[0]);r.parts[0]=r.parts[0][n:];if r.parts[0]==""{r.parts=r.parts[1:]};if len(r.parts)==0{return n,r.final};return n,nil
}
func(r *reviewExportReader)Close()error{r.closes++;return nil}
type reviewExportWriter struct{buf bytes.Buffer;writes,closes int;writeErr,closeErr error;short bool}
func(w *reviewExportWriter)Write(p []byte)(int,error){w.writes++;if w.writeErr!=nil{return 0,w.writeErr};if w.short{return len(p)-1,nil};return w.buf.Write(p)}
func(w *reviewExportWriter)Close()error{w.closes++;return w.closeErr}
type reviewExportCause struct{where string}
func(e *reviewExportCause)Error()string{return e.where}
func TestIndependentReviewExportCompletion(t *testing.T){
 readErr:=&reviewExportCause{"read"};writeErr:=&reviewExportCause{"write"};closeErr:=&reviewExportCause{"close"}
 for _,tc:=range []struct{name string;readErr,writeErr,closeErr error;want int}{
 {"read-only",readErr,nil,nil,2},{"close-only",io.EOF,nil,closeErr,2},{"both read-close",readErr,nil,closeErr,2},{"write-only",io.EOF,writeErr,nil,0},{"write-close",io.EOF,writeErr,closeErr,0},{"success",io.EOF,nil,nil,2},
 }{t.Run(tc.name,func(t *testing.T){
 w:=&reviewExportWriter{writeErr:tc.writeErr,closeErr:tc.closeErr};r:=&reviewExportReader{parts:[]string{"a\n","last"},final:tc.readErr};n,err:=lineexport.Export(r,w)
 if n!=tc.want||w.closes!=1||r.closes!=0{t.Fatalf("n=%d closes=%d reader closes=%d err=%v",n,w.closes,r.closes,err)}
 primary:=tc.writeErr;if primary==nil&&tc.readErr!=io.EOF{primary=tc.readErr}
 if primary==nil {if err!=tc.closeErr{t.Fatalf("single close identity=%v",err)}} else if tc.closeErr==nil {if err!=primary{t.Fatalf("single primary identity=%v",err)}} else {if !errors.Is(err,primary)||!errors.Is(err,tc.closeErr){t.Fatalf("lost joined identity=%v",err)};var typed *reviewExportCause;if !errors.As(err,&typed){t.Fatalf("lost error type=%v",err)}}
 if tc.want==2&&w.buf.String()!="a\nlast\n"{t.Fatalf("output=%q",w.buf.String())}
 })}
}
func TestIndependentReviewExportLimitAndOwnership(t *testing.T){
 for _,limit:=range []int{-1,1,2,0}{w:=new(reviewExportWriter);r:=&reviewExportReader{parts:[]string{"a\n","\n"},written:&w.writes,checkOrder:true,final:io.EOF};n,err:=lineexport.ExportLimit(r,w,limit)
 if w.closes!=1||r.closes!=0{t.Fatal("ownership")}
 if limit<0 {if n!=0||err==nil||r.reads!=0||w.writes!=0{t.Fatalf("negative n=%d err=%v reads=%d writes=%d",n,err,r.reads,w.writes)};continue}
 want:=2;if limit==1{want=1};if n!=want||err!=nil{t.Fatalf("limit=%d n=%d err=%v",limit,n,err)}
 if limit==1&&r.reads!=1{t.Fatalf("limited export read extra chunk: %d",r.reads)}
 }
 w:=&reviewExportWriter{short:true};r:=&reviewExportReader{parts:[]string{"a\n"},final:io.EOF};n,err:=lineexport.Export(r,w)
 if n!=0||err!=io.ErrShortWrite||w.closes!=1{t.Fatalf("short write n=%d err=%v closes=%d",n,err,w.closes)}
}
