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

func TestCLI(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "lineexport")
	buildCtx, cancelBuild := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelBuild()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binary, ".")
	build.Env = os.Environ()
	build.Env = append(build.Env, "GOWORK=off", "GOTOOLCHAIN=local")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	for _, tc := range []struct {
		name       string
		args       []string
		input      string
		wantStatus int
		wantOutput string
	}{
		{"unlimited", []string{"-input", "-output"}, "a\n\nlast", 0, "a\n\nlast\n"},
		{"zero", []string{"-input", "-output", "0"}, "a\nb\n", 0, "a\nb\n"},
		{"prefix", []string{"-input", "-output", "1"}, "a\nb\n", 0, "a\n"},
		{"above count", []string{"-input", "-output", "3"}, "a\nb\n", 0, "a\nb\n"},
		{"missing arguments", []string{"-input"}, "a\n", 2, "original"},
		{"extra arguments", []string{"-input", "-output", "1", "extra"}, "a\n", 2, "original"},
		{"invalid limit", []string{"-input", "-output", "one"}, "a\n", 2, "original"},
		{"negative limit", []string{"-input", "-output", "-1"}, "a\n", 2, "original"},
		{"overflow limit", []string{"-input", "-output", "999999999999999999999999999999999999"}, "a\n", 2, "original"},
		{"missing input", []string{"-missing", "-output"}, "a\n", 1, "original"},
		{"cannot create output", []string{"-input", "missing/output"}, "a\n", 1, "original"},
		{"read failure retains prefix", []string{"-input", "-output"}, "a\n" + strings.Repeat("x", 70<<10), 1, "a\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "-input"), []byte(tc.input), 0600); err != nil {
				t.Fatal(err)
			}
			outputPath := filepath.Join(dir, "-output")
			if err := os.WriteFile(outputPath, []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, tc.args...)
			cmd.Dir = dir
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			status := 0
			if err != nil {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("run CLI: %v", err)
				}
				status = exitErr.ExitCode()
			}
			if ctx.Err() != nil {
				t.Fatalf("CLI did not finish: %v", ctx.Err())
			}
			if status != tc.wantStatus {
				t.Errorf("exit = %d, want %d; stderr = %q", status, tc.wantStatus, stderr.String())
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			if (tc.wantStatus == 0) != (stderr.Len() == 0) {
				t.Errorf("stderr = %q, want empty only on success", stderr.String())
			}
			output, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(output) != tc.wantOutput {
				t.Errorf("output = %q, want %q", output, tc.wantOutput)
			}
		})
	}
}
