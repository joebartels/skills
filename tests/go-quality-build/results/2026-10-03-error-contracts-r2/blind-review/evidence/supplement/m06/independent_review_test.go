package entryzip_test
import(
 "archive/zip"
 "bytes"
 "errors"
 "io"
 "math/rand"
 "strings"
 "testing"
 "example.com/entryzip"
)
type reviewZipBody struct{data []byte;cause error;reads,closes int;terminal *bool}
func(r *reviewZipBody)Read(p []byte)(int,error){r.reads++;n:=copy(p,r.data);r.data=r.data[n:];if len(r.data)==0{if r.terminal!=nil{*r.terminal=true};if r.cause!=nil{return n,r.cause};return n,io.EOF};return n,nil}
func(r *reviewZipBody)Close()error{r.closes++;return nil}
type reviewZipOutput struct{buf bytes.Buffer;cause error;writes,closes int;terminal *bool}
func(w *reviewZipOutput)Write(p []byte)(int,error){w.writes++;if w.cause!=nil&&(w.terminal==nil||*w.terminal){return 0,w.cause};return w.buf.Write(p)}
func(w *reviewZipOutput)Close()error{w.closes++;return nil}
type reviewZipCause struct{where string}
func(e *reviewZipCause)Error()string{return e.where}
type reviewZipNonComparable []string
func(e reviewZipNonComparable)Error()string{return strings.Join(e," ")}
func reviewZipContents(t *testing.T,data []byte,names,bodies []string){
 t.Helper();r,err:=zip.NewReader(bytes.NewReader(data),int64(len(data)));if err!=nil{t.Fatalf("central directory missing: %v",err)}
 if len(r.File)!=len(names){t.Fatalf("entry count=%d want=%d",len(r.File),len(names))}
 for i,f:=range r.File{if f.Name!=names[i]{t.Fatalf("entry %d=%q want=%q",i,f.Name,names[i])};body,err:=f.Open();if err!=nil{t.Fatal(err)};b,readErr:=io.ReadAll(body);closeErr:=body.Close();if readErr!=nil||closeErr!=nil||string(b)!=bodies[i]{t.Fatalf("entry %d body=%q read=%v close=%v",i,b,readErr,closeErr)}}
}
func TestIndependentReviewZipFilteringAndOwnership(t *testing.T){
 output:=new(reviewZipOutput);skip:=&reviewZipBody{data:[]byte("never read")};first:=&reviewZipBody{data:[]byte("one")};last:=&reviewZipBody{data:[]byte("two")};names:=[]string{"z","z"};calls:=[]string{}
 err:=entryzip.WriteSelected(output,[]entryzip.Entry{{Name:"z",Body:first},{Name:strings.Repeat("x",65536),Body:skip},{Name:"z",Body:last}},func(name string)bool{calls=append(calls,name);return name=="z"})
 if err!=nil{t.Fatal(err)};reviewZipContents(t,output.buf.Bytes(),names,[]string{"one","two"})
 if len(calls)!=3||calls[0]!="z"||len(calls[1])!=65536||calls[2]!="z"||skip.reads!=0{t.Fatalf("filter calls/read boundary=%v/%d",len(calls),skip.reads)}
 if first.closes!=0||skip.closes!=0||last.closes!=0||output.closes!=0{t.Fatal("borrowed resource closed")}
 output=new(reviewZipOutput);if err:=entryzip.WriteSelected(output,[]entryzip.Entry{{Name:"skip",Body:nil}},func(string)bool{return false});err!=nil{t.Fatal(err)};reviewZipContents(t,output.buf.Bytes(),nil,nil)
}
func TestIndependentReviewZipCompletion(t *testing.T){
 bodyErr:=&reviewZipCause{"body"};finishErr:=&reviewZipCause{"completion"}
 for _,tc:=range []struct{name string;body,finish error}{{"success",nil,nil},{"body",bodyErr,nil},{"finish",nil,finishErr},{"both",bodyErr,finishErr}}{
 t.Run(tc.name,func(t *testing.T){terminal:=false;body:=&reviewZipBody{data:[]byte("prefix"),cause:tc.body,terminal:&terminal};later:=&reviewZipBody{data:[]byte("later")};output:=&reviewZipOutput{cause:tc.finish,terminal:&terminal}
 err:=entryzip.WriteSelected(output,[]entryzip.Entry{{Name:"first",Body:body},{Name:"later",Body:later}},nil)
 switch{case tc.body!=nil&&tc.finish!=nil:if !errors.Is(err,tc.body)||!errors.Is(err,tc.finish){t.Fatalf("lost independent completion=%v",err)};var typed *reviewZipCause;if !errors.As(err,&typed){t.Fatal("typed body cause hidden")}
 case tc.body!=nil:if err!=tc.body{t.Fatalf("lone body identity=%v",err)}
 case tc.finish!=nil:if err!=tc.finish{t.Fatalf("lone completion identity=%v",err)}
 default:if err!=nil{t.Fatal(err)};reviewZipContents(t,output.buf.Bytes(),[]string{"first","later"},[]string{"prefix","later"})}
 if body.closes!=0||later.closes!=0||output.closes!=0{t.Fatal("borrowed resource closed")};if tc.body!=nil&&later.reads!=0{t.Fatal("continued past body failure")};if output.writes!=1{t.Fatalf("small archive expected one deferred flush, writes=%d",output.writes)}
 })}
 output:=&reviewZipOutput{cause:finishErr};if err:=entryzip.WriteArchive(output,nil);err!=finishErr{t.Fatalf("empty archive completion identity=%v",err)}
}
func TestIndependentReviewZipStickyOutputIdentity(t *testing.T){
 cause:=reviewZipNonComparable{"caller","output"};large:=make([]byte,256*1024);_,_ = rand.New(rand.NewSource(31)).Read(large)
 for _,name:=range []string{"copy","create"}{t.Run(name,func(t *testing.T){
  output:=&reviewZipOutput{cause:cause};body:=&reviewZipBody{data:large};entryName:="data";if name=="create"{entryName=strings.Repeat("x",5000)}
  err:=entryzip.WriteArchive(output,[]entryzip.Entry{{Name:entryName,Body:body}})
  direct,ok:=err.(reviewZipNonComparable);if !ok||len(direct)!=len(cause)||&direct[0]!=&cause[0]{t.Fatalf("lone caller output wrapped or changed: %T %v",err,err)}
  if output.writes!=1||output.closes!=0||body.closes!=0{t.Fatalf("writes=%d output closes=%d body closes=%d",output.writes,output.closes,body.closes)}
  if name=="create"&&body.reads!=0{t.Fatal("body read after creation output failure")}
 })}
}
func TestIndependentReviewZipCreationAndCompletion(t *testing.T){
 finish:=&reviewZipCause{"finish"};output:=&reviewZipOutput{cause:finish};body:=&reviewZipBody{data:[]byte("unread")};err:=entryzip.WriteSelected(output,[]entryzip.Entry{{Name:strings.Repeat("x",65536),Body:body}},nil)
 if err==nil||!errors.Is(err,finish)||err==finish{t.Fatalf("creation/finish error=%v",err)}
 if joined,ok:=err.(interface{Unwrap()[]error});!ok||len(joined.Unwrap())!=2{t.Fatalf("independent creation cause missing: %T",err)}
 if body.reads!=0||body.closes!=0||output.closes!=0||output.writes!=1{t.Fatalf("creation ownership reads=%d closes=%d writes=%d",body.reads,body.closes,output.writes)}
}

func TestIndependentReviewStandardZipStickyEvidence(t *testing.T){
 cause:=&reviewZipCause{"output"};output:=&reviewZipOutput{cause:cause};archive:=zip.NewWriter(output)
 _,primary:=archive.Create(strings.Repeat("x",5000));finish:=archive.Close()
 if primary!=cause||finish!=cause||output.writes!=1{t.Fatalf("standard ZIP sticky observation primary=%v finish=%v writes=%d",primary,finish,output.writes)}
 t.Log("one underlying caller output failure reappears from standard ZIP Close; preserving lone identity needs deduplication")
}
