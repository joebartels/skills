package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jobs.json")
	if err := os.WriteFile(path, []byte(`[{"id":"build","state":"ready"}]`), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	if code := run([]string{path}, &out, &diag); code != 0 {
		t.Fatalf("exit %d: %s", code, &diag)
	}
	if got := out.String(); got != "{\"jobs\":[{\"id\":\"build\",\"state\":\"ready\"}]}\n" {
		t.Fatalf("output = %q", got)
	}
}
