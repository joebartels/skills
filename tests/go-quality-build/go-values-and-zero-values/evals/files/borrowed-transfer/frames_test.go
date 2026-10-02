package framecollect

import (
	"io"
	"testing"
)

type empty struct{}

func (empty) Next() ([]byte, error) { return nil, io.EOF }
func TestEmpty(t *testing.T) {
	b, err := Collect(empty{})
	if len(b) != 0 || err != nil {
		t.Fatalf("%v %v", b, err)
	}
}
