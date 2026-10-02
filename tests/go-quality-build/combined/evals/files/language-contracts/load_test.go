package batchview

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	got, e := Load(strings.NewReader("a=one\nb=two\n"))
	if len(got) != 2 || e != nil {
		t.Fatalf("%v %v", got, e)
	}
}
