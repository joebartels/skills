package csvselect

import (
	"bytes"
	"testing"
)

func TestWriteRows(t *testing.T) {
	var w bytes.Buffer
	if err := WriteRows(&w, [][]string{{"a", "b"}}); err != nil || w.String() != "a,b\n" {
		t.Fatalf("%q %v", w.String(), err)
	}
}
