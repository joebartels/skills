package recordload
import("errors";"io";"strings";"testing")
type onceReader struct{data string;err error;done bool}
func(r *onceReader)Read(p []byte)(int,error){if r.done{return 0,io.EOF};r.done=true;return copy(p,r.data),r.err}
func TestContractReaderDataAndFailure(t *testing.T){cause:=errors.New("reader failure");got,err:=Load(&onceReader{data:"a=1\nb=2",err:cause});if len(got)!=2||!errors.Is(err,cause){t.Fatalf("lost data/cause: %#v %v",got,err)}}
func TestContractStructuredLineAndSuccess(t *testing.T){got,err:=Load(strings.NewReader("# ignored\n\na=1\nbad\n"));var line *LineError;if len(got)!=1||!errors.As(err,&line)||line.Line!=4||line.Text!="bad"{t.Fatalf("%#v %v",got,err)};got,err=Load(strings.NewReader("a=1\n"));if err!=nil||len(got)!=1{t.Fatalf("success error: %#v %v",got,err)}}
func TestContractLongLineAndEOF(t *testing.T){v:=strings.Repeat("v",100000);got,err:=Load(strings.NewReader("a="+v));if err!=nil||len(got)!=1||got[0].Value!=v{t.Fatalf("long record: %d %v",len(got),err)}}

func TestContractReadAndParseFailure(t *testing.T) {
    cause := errors.New("reader failed after delivering bytes")
    got, err := Load(&onceReader{data: "# comment\nok=1\nmalformed\n", err: cause})
    var line *LineError
    if len(got) != 1 || got[0].Key != "ok" || got[0].Value != "1" || !errors.Is(err, cause) || !errors.As(err, &line) || line.Line != 3 || line.Text != "malformed" {
        t.Fatalf("lost prefix or simultaneous cause: %#v %v", got, err)
    }
}
