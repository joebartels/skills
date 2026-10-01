package span_test

import (
	"example.com/span"
	"testing"
)

var constructor func(int, int) (span.Range, error) = span.New

func TestContract(t *testing.T) {
	r, err := constructor(2, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Contains(2) || r.Contains(4) {
		t.Fatal("half-open contract")
	}
	if _, err := span.New(4, 2); err == nil {
		t.Fatal("accepted reversed range")
	}
	if (span.Range{}).Contains(0) {
		t.Fatal("zero range must be empty")
	}
}
