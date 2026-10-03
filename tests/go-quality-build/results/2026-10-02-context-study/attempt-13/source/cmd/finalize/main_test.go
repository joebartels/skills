package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func buildCommand(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "finalize")
	cmd := exec.Command("go", "build", "-o", binary, "./cmd/finalize")
	cmd.Dir = "../.."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build command: %v\n%s", err, out)
	}
	return binary
}

func TestInterruptedCommandFinalizesAcceptedReceipt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Interrupt delivery to child processes is unsupported on Windows")
	}
	binary := buildCommand(t)
	receipt := filepath.Join(t.TempDir(), "receipt")
	cmd := exec.Command(binary, "--receipt", receipt, "1", "0")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	finished := false
	t.Cleanup(func() {
		if !finished && cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})
	lines := bufio.NewReader(stderr)
	type stderrResult struct {
		firstLine string
		rest      []byte
		err       error
	}
	type readyResult struct {
		firstLine string
		err       error
	}
	stderrDone := make(chan stderrResult, 1)
	ready := make(chan readyResult, 1)
	go func() {
		line, err := lines.ReadString('\n')
		if err == nil && strings.TrimSpace(line) != "waiting" {
			err = io.ErrUnexpectedEOF
		}
		ready <- readyResult{firstLine: line, err: err}
		rest, readErr := io.ReadAll(lines)
		if err == nil {
			err = readErr
		}
		stderrDone <- stderrResult{firstLine: line, rest: rest, err: err}
	}()
	select {
	case got := <-ready:
		if got.err != nil || strings.TrimSpace(got.firstLine) != "waiting" {
			t.Fatalf("waiting signal: line=%q err=%v", got.firstLine, got.err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("command did not report waiting (stdout=%q)", stdout.String())
	}
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("interrupt command: %v", err)
	}
	waitErr := cmd.Wait()
	finished = true
	gotStderr := <-stderrDone
	if gotStderr.err != nil {
		t.Fatalf("read command stderr: %v", gotStderr.err)
	}
	if exit, ok := waitErr.(*exec.ExitError); !ok || exit.ExitCode() != 2 {
		t.Fatalf("command status=%v, want exit 2 (stdout=%q stderr=%q)", waitErr, stdout.String(), gotStderr.rest)
	}
	if got := stdout.String(); got != "accepted 1\n" {
		t.Errorf("stdout=%q, want accepted output", got)
	}
	if got := string(gotStderr.rest); !strings.Contains(got, "context canceled") {
		t.Errorf("stderr=%q, want cancellation diagnostic", got)
	}
	contents, err := os.ReadFile(receipt)
	if err != nil {
		t.Fatalf("read receipt: %v", err)
	}
	if string(contents) != "accepted=1\n" {
		t.Errorf("receipt=%q, want accepted=1 newline", contents)
	}
}

func TestCommandReportsReceiptWriteFailure(t *testing.T) {
	binary := buildCommand(t)
	receiptDir := t.TempDir()
	cmd := exec.Command(binary, "--receipt", receiptDir, "1")
	stdout, err := cmd.Output()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 2 {
		t.Fatalf("command error=%v, want exit 2 (stdout=%q)", err, stdout)
	}
	if string(stdout) != "accepted 1\n" {
		t.Errorf("stdout=%q, want accepted output", stdout)
	}
	if !bytes.Contains(exit.Stderr, []byte(receiptDir)) {
		t.Errorf("stderr=%q, want receipt path in write failure", exit.Stderr)
	}
}
