package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var executable string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "indexer-process-tests-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	executable = filepath.Join(dir, "indexer")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	command := exec.CommandContext(ctx, "go", "build", "-o", executable, ".")
	output, err := command.CombinedOutput()
	cancel()
	if err != nil {
		os.RemoveAll(dir)
		fmt.Fprintf(os.Stderr, "build command: %v\n%s", err, output)
		os.Exit(1)
	}
	status := m.Run()
	os.RemoveAll(dir)
	os.Exit(status)
}

func TestExistingPut(t *testing.T) {
	if err := run([]string{"put", t.TempDir(), "alpha", "one"}); err != nil {
		t.Fatal(err)
	}
	if err := run(nil); err == nil {
		t.Fatal("missing usage rejection")
	}
}

func TestPutProcessContract(t *testing.T) {
	workingDir := t.TempDir()
	for _, text := range []string{"first", "", "-literal text"} {
		result := invoke(t, workingDir, "", "put", "-index", "alpha", text)
		assertProcess(t, result, 0)
		assertBytes(t, filepath.Join(workingDir, "-index", "alpha"), text)
	}
	result := invoke(t, workingDir, "", "put", "-index", "Alpha", "rejected")
	assertProcess(t, result, 2)
	assertBytes(t, filepath.Join(workingDir, "-index", "alpha"), "-literal text")
}

func TestApplyProcessContract(t *testing.T) {
	dir := t.TempDir()
	result := invoke(t, t.TempDir(), "alpha,first\nbeta,\"two, words\"\nalpha,last\n", "apply", "--dir", dir)
	assertProcess(t, result, 0)
	assertBytes(t, filepath.Join(dir, "alpha"), "last")
	assertBytes(t, filepath.Join(dir, "beta"), "two, words")
	for _, bad := range []string{"alpha,\" \t \"\n", "alpha,bad,extra\n", "alpha,\"unterminated\n"} {
		result := invoke(t, t.TempDir(), "alpha,accepted\n"+bad+"gamma,later\n", "apply", "--dir", dir)
		assertProcess(t, result, 2)
		assertBytes(t, filepath.Join(dir, "alpha"), "accepted")
		if _, err := os.Stat(filepath.Join(dir, "gamma")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("later row exists: %v", err)
		}
	}
}

func TestUsageProcessContract(t *testing.T) {
	for _, args := range [][]string{
		nil, {"put"}, {"unknown"}, {"apply"}, {"apply", "--dir", t.TempDir(), "extra"},
		{"watch", "--url", "ftp://example.test", "--file", "out", "--interval", "1s"},
		{"watch", "--url", "http:///missing", "--file", "out", "--interval", "1s"},
		{"watch", "--url", "https://example.test", "--file", "out", "--interval", "0s"},
		{"watch", "--url", "https://example.test", "--file", "out", "--interval", "-1s"},
		{"watch", "--url", "https://example.test", "--file", "out", "--interval", "no"},
		{"watch", "--url", "https://example.test", "--interval", "1s"},
		{"watch", "--url", "https://example.test", "--file", "out", "--interval", "1s", "extra"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) { assertProcess(t, invoke(t, t.TempDir(), "", args...), 2) })
	}
}

type processResult struct {
	status         int
	stdout, stderr string
}

func invoke(t *testing.T, cwd, input string, args ...string) processResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = cwd
	command.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("process deadline: %v; stderr %s", ctx.Err(), stderr.String())
	}
	status := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("execute: %v", err)
		}
		status = exitErr.ExitCode()
	}
	return processResult{status, stdout.String(), stderr.String()}
}
func assertProcess(t *testing.T, result processResult, status int) {
	t.Helper()
	if result.status != status || result.stdout != "" || (status == 0 && result.stderr != "") || (status != 0 && result.stderr == "") {
		t.Fatalf("process = status %d, stdout %q, stderr %q; want status %d", result.status, result.stdout, result.stderr, status)
	}
}
func assertBytes(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file = %q, %v; want %q", got, err, want)
	}
}
