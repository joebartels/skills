package lineexport

import (
	"bytes"
	"strings"
	"testing"
)

type bufferCloser struct{ bytes.Buffer }

func (*bufferCloser) Close() error { return nil }
func TestExport(t *testing.T) {
	w := new(bufferCloser)
	n, err := Export(strings.NewReader("a\nb\n"), w)
	if n != 2 || err != nil || w.String() != "a\nb\n" {
		t.Fatalf("%d %v %q", n, err, w.String())
	}
}
