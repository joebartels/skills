package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	p := filepath.Join(t.TempDir(), "numbers")
	if err := os.WriteFile(p, []byte(" 2\n\n-1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	if err := run([]string{p}, &b); err != nil {
		t.Fatal(err)
	}
	if b.String() != "1\n" {
		t.Fatalf("got %q", b.String())
	}
}
func TestFailureHasNoOutput(t *testing.T) {
	p := filepath.Join(t.TempDir(), "numbers")
	if err := os.WriteFile(p, []byte("2\nnope\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var b bytes.Buffer
	if run([]string{p}, &b) == nil || b.Len() != 0 {
		t.Fatalf("invalid input accepted or output written: %q", b.String())
	}
}
