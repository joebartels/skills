package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSingleRecord(t *testing.T) {
	dir := t.TempDir()
	if err := run([]string{"--dir", dir}, strings.NewReader("web,7\n"), io.Discard); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "web"))
	if err != nil || string(got) != "7\n" {
		t.Fatalf("record = %q, %v", got, err)
	}
}
