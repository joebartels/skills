package recordload
import("errors";"testing";"io")
func TestReviewSimultaneousMalformedReadFailure(t *testing.T) { cause:=errors.New("underlying read failure"); got,err:=Load(&onceReader{data:"a=1\nbroken",err:cause}); var line *LineError; if len(got)!=1 || !errors.As(err,&line) || !errors.Is(err,cause) { t.Fatalf("prefix=%#v err=%v line=%v readCause=%v",got,err,line,errors.Is(err,cause)) } }
func TestReviewBytesWithEOF(t *testing.T) { got,err:=Load(&onceReader{data:"=\n#hi\nx=y",err:io.EOF});if len(got)!=2||err!=nil {t.Fatal(got,err)} }
