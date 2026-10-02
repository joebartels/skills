package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRejectedBatchProcess(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "ledgerload")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(ctx, "go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, out)
	}
	for _, input := range []string{"first,7\nbad,zero\nlast,9\n", "first,7\nbad,\"unterminated\nlast,9\n"} {
		t.Run(input, func(t *testing.T) {
			dir := t.TempDir()
			cmd := exec.CommandContext(ctx, bin, "--dir", dir)
			cmd.Stdin = strings.NewReader(input)
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 2 || stdout.Len() != 0 || stderr.Len() == 0 {
				t.Errorf("process = %v; stdout=%q stderr=%q; want exit2 diagnostic", err, stdout.String(), stderr.String())
			}
			first, err := os.ReadFile(filepath.Join(dir, "first"))
			if err != nil || string(first) != "7\n" {
				t.Errorf("accepted prefix = %q, %v", first, err)
			}
			if _, err := os.Stat(filepath.Join(dir, "last")); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("later record was not absent: %v", err)
			}
		})
	}
	dir := t.TempDir()
	cmd := exec.CommandContext(ctx, bin, "--dir", dir)
	cmd.Stdin = strings.NewReader("web,4\nweb,8\napi,9\n")
	if out, err := cmd.CombinedOutput(); err != nil || len(out) != 0 {
		t.Fatalf("success process: %v, %q", err, out)
	}
	for id, want := range map[string]string{"web": "8\n", "api": "9\n"} {
		got, err := os.ReadFile(filepath.Join(dir, id))
		if err != nil || string(got) != want {
			t.Errorf("%s = %q, %v; want %q", id, got, err, want)
		}
	}
}
