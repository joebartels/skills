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

func TestLineexport(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "lineexport")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}

	for _, tc := range []struct {
		name       string
		args       []string
		input      string
		dashPaths  bool
		outputDir  bool
		wantExit   int
		wantOutput string
	}{
		{"unlimited", []string{"input", "output"}, "a\n\nlast", false, false, 0, "a\n\nlast\n"},
		{"zero", []string{"input", "output", "0"}, "a\n\nlast", false, false, 0, "a\n\nlast\n"},
		{"positive", []string{"input", "output", "1"}, "a\n\nlast", false, false, 0, "a\n"},
		{"exact count", []string{"input", "output", "3"}, "a\n\nlast", false, false, 0, "a\n\nlast\n"},
		{"above count", []string{"input", "output", "4"}, "a\n\nlast", false, false, 0, "a\n\nlast\n"},
		{"dash paths", []string{"-input", "-output"}, "a\nb", true, false, 0, "a\nb\n"},
		{"dash paths with limit", []string{"-input", "-output", "1"}, "a\nb", true, false, 0, "a\n"},
		{"no arguments", nil, "a\n", false, false, 2, "previous output\n"},
		{"one argument", []string{"input"}, "a\n", false, false, 2, "previous output\n"},
		{"too many arguments", []string{"input", "output", "1", "extra"}, "a\n", false, false, 2, "previous output\n"},
		{"malformed limit", []string{"input", "output", "one"}, "a\n", false, false, 2, "previous output\n"},
		{"negative limit", []string{"input", "output", "-1"}, "a\n", false, false, 2, "previous output\n"},
		{"overflowing limit", []string{"input", "output", "999999999999999999999999999"}, "a\n", false, false, 2, "previous output\n"},
		{"missing input", []string{"missing", "output"}, "a\n", false, false, 1, "previous output\n"},
		{"output is directory", []string{"input", "output"}, "a\n", false, true, 1, ""},
		{"read failure after prefix", []string{"input", "output"}, "first\n" + strings.Repeat("x", 70*1024) + "\nlast\n", false, false, 1, "first\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			inputName, outputName := "input", "output"
			if tc.dashPaths {
				inputName, outputName = "-input", "-output"
			}
			if err := os.WriteFile(filepath.Join(dir, inputName), []byte(tc.input), 0600); err != nil {
				t.Fatal(err)
			}
			outputPath := filepath.Join(dir, outputName)
			if tc.outputDir {
				if err := os.Mkdir(outputPath, 0700); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(outputPath, []byte("previous output\n"), 0600); err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, tc.args...)
			cmd.Dir = dir
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			if ctx.Err() != nil {
				t.Fatalf("CLI timed out: %v", ctx.Err())
			}
			exit := 0
			if err != nil {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("run CLI: %v", err)
				}
				exit = exitErr.ExitCode()
			}
			if exit != tc.wantExit {
				t.Errorf("exit status = %d, want %d; stderr %q", exit, tc.wantExit, stderr.String())
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
			if (tc.wantExit == 0 && stderr.Len() != 0) || (tc.wantExit != 0 && stderr.Len() == 0) {
				t.Errorf("stderr = %q, want empty only on success", stderr.String())
			}
			if tc.wantExit == 2 && !strings.Contains(stderr.String(), "usage:") {
				t.Errorf("invalid usage stderr = %q, want usage", stderr.String())
			}
			if tc.outputDir {
				info, err := os.Stat(outputPath)
				if err != nil || !info.IsDir() {
					t.Errorf("output directory changed: %v", err)
				}
				return
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
