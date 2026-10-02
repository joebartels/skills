package recordload
import("errors";"testing";"io")
type reviewOnce struct{done bool; cause error}
func(r *reviewOnce)Read(p []byte)(int,error){if r.done{return 0,io.EOF};r.done=true;return copy(p,"a=1\nbroken"),r.cause}
func TestReviewBaselineCause(t *testing.T){cause:=errors.New("underlying read failure");got,err:=Load(&reviewOnce{cause:cause});t.Logf("prefix=%#v err=%v readCause=%v",got,err,errors.Is(err,cause));if !errors.Is(err,cause){t.Fatal("cause lost")}}
