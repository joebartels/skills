package csvselect
import("strings";"errors";"testing";"io")
func TestReviewImmediateCSVFailure(t *testing.T){cause:=errors.New("immediate sink failure");w:=&LCFailCSVWriter{cause:cause};n,e:=WriteSelected(w,[][]string{{strings.Repeat("x",5000)},{"later"}},nil);if n!=1||!errors.Is(e,cause)||w.closed {t.Fatal(n,e,w.closed)};var _ func(io.Writer,[][]string)error=WriteRows}
