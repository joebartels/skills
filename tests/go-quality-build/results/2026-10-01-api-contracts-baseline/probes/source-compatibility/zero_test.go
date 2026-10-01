package consumer_test
import("testing"; "example.com/recordfmt")
func TestZeroValue(t *testing.T){ var f recordfmt.Formatter; if got:=f.Format("key","value");got!="key:value" { t.Fatalf("zero Format = %q",got) } }
