package lineexport
import("errors";"testing";"strings";"io")
func TestReviewLimitReadFailureObservation(t *testing.T) {cause:=errors.New("read failure with accepted bytes");w:=new(probeWriter);n,e:=ExportLimit(&probeFailReader{cause:cause},w,1);t.Logf("limit=1 n=%d err=%v closed=%d output=%q",n,e,w.closed,w.String());if n!=1||w.closed!=1||w.String()!="a\n" {t.Fatal("lost accepted prefix")} }
func TestReviewFunctionType(t *testing.T){var _ func(io.Reader,io.WriteCloser)(int,error)=Export;w:=new(probeWriter);n,e:=ExportLimit(strings.NewReader("\nlast"),w,0);if n!=2||e!=nil||w.String()!="\nlast\n" {t.Fatal(n,e,w.String())}}
