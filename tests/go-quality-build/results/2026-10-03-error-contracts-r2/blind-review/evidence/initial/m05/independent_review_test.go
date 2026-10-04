package recordload_test

import (
 "errors"
 "fmt"
 "io"
 "reflect"
 "strings"
 "testing"
 "example.com/recordload"
)

type reviewReadCause struct{ label string }
func (e *reviewReadCause) Error() string { return e.label }
type reviewChunkReader struct { data string; step int; final error; closes int }
func (r *reviewChunkReader) Read(p []byte) (int,error) {
 if r.data == "" { return 0,r.final }
 n:=len(p); if n>r.step {n=r.step}; if n>len(r.data) {n=len(r.data)}
 copy(p,r.data[:n]); r.data=r.data[n:]
 if r.data=="" { return n,r.final }; return n,nil
}
func (r *reviewChunkReader) Close() error {r.closes++;return nil}

func TestIndependentReviewLoaderBoundaries(t *testing.T) {
 cause := &reviewReadCause{"read boundary"}
 for _,tc := range []struct{name,data string; final error; want []recordload.Record; line int; text string}{
  {"complete literal records", "#x\n\n=\na=b=c\n a = b\r\n", io.EOF, []recordload.Record{{},{Key:"a",Value:"b=c"},{Key:" a ",Value:" b\r"}},0,""},
  {"malformed and typed read error", "#x\n\na=1\n bad \nz=2", fmt.Errorf("source: %w",cause), []recordload.Record{{Key:"a",Value:"1"}},4," bad "},
  {"valid unterminated with typed read error", "a=1\nb=2",cause,[]recordload.Record{{Key:"a",Value:"1"},{Key:"b",Value:"2"}},0,""},
  {"EOF with malformed suffix", "a=1\n\nwrong",io.EOF,[]recordload.Record{{Key:"a",Value:"1"}},3,"wrong"},
  {"empty success", "",io.EOF,nil,0,""},
 } {
  t.Run(tc.name,func(t *testing.T){
   r:=&reviewChunkReader{data:tc.data,step:1,final:tc.final}; got,err:=recordload.Load(r)
   if !reflect.DeepEqual(got,tc.want){t.Fatalf("records=%#v want=%#v",got,tc.want)}
   if r.closes!=0 {t.Fatal("closed borrowed reader")}
   var line *recordload.LineError
   if tc.line!=0 {if !errors.As(err,&line)||line.Line!=tc.line||line.Text!=tc.text {t.Fatalf("line error=%#v err=%v",line,err)}} else if errors.As(err,&line){t.Fatalf("unexpected line error=%#v",line)}
   if tc.final==io.EOF {
    if errors.Is(err,io.EOF) {t.Fatal("EOF exposed")}
    if tc.line==0&&err!=nil {t.Fatalf("success error interface=%#v",err)}
   } else {
    var gotCause *reviewReadCause
    if !errors.Is(err,cause)||!errors.As(err,&gotCause)||gotCause!=cause {t.Fatalf("lost original cause: %v",err)}
   }
  })
 }
 value:=strings.Repeat("x",256*1024); got,err:=recordload.Load(strings.NewReader("key="+value))
 if err!=nil||len(got)!=1||got[0].Value!=value {t.Fatalf("long record count=%d err=%v",len(got),err)}
}
