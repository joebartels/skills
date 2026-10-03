package main

import (
	"bufio"
	"bytes"
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
	path := filepath.Join(t.TempDir(), "finalize")
	if runtime.GOOS == "windows" {
		path += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", path, ".")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build command: %v: %s", err, stderr.String())
	}
	return path
}

func TestCommandCompletesAndWritesReceipt(t *testing.T) {
	bin := buildCommand(t)
	receipt := filepath.Join(t.TempDir(), "receipt")
	cmd := exec.Command(bin, "--receipt", receipt, "3", "4")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("command: %v stderr=%q", err, stderr.String())
	}
	if got := stdout.String(); got != "accepted 3\naccepted 4\n" {
		t.Fatalf("stdout=%q", got)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr=%q", got)
	}
	data, err := os.ReadFile(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "accepted=2\n" {
		t.Fatalf("receipt=%q", data)
	}
}

func TestCommandReceiptWriteFailure(t *testing.T) {
	bin := buildCommand(t)
	missing := filepath.Join(t.TempDir(), "missing", "receipt")
	cmd := exec.Command(bin, "--receipt", missing, "8")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("command succeeded despite receipt write failure")
	}
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 2 {
		t.Fatalf("exit=%v", err)
	}
	if stdout.String() != "accepted 8\n" {
		t.Fatalf("stdout=%q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "no such file or directory") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestCommandInterruptFinalizesAcceptedPrefix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Interrupt cannot be delivered to child processes on Windows")
	}
	bin := buildCommand(t)
	receipt := filepath.Join(t.TempDir(), "receipt")
	cmd := exec.Command(bin, "--receipt", receipt, "7", "0")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	lines := make(chan string, 2)
	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			lines <- s.Text()
		}
		close(lines)
	}()
	readOutput := make(chan string, 1)
	go func() { var b bytes.Buffer; _, _ = b.ReadFrom(stdout); readOutput <- b.String() }()
	seenWaiting := false
	deadline := time.After(5 * time.Second)
	for !seenWaiting {
		select {
		case line, ok := <-lines:
			if !ok {
				t.Fatal("command exited before waiting")
			}
			if line == "waiting" {
				seenWaiting = true
			}
		case <-deadline:
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			t.Fatal("command did not reach cooperative wait")
		}
	}
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("interrupt: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 2 {
			t.Fatalf("exit=%v", err)
		}
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		<-done
		t.Fatal("command did not finalize after interruption")
	}
	select {
	case got := <-readOutput:
		if got != "accepted 7\n" {
			t.Fatalf("stdout=%q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("stdout did not close")
	}
	data, err := os.ReadFile(receipt)
	if err != nil {
		t.Fatalf("receipt after interruption: %v", err)
	}
	if string(data) != "accepted=1\n" {
		t.Fatalf("receipt=%q", data)
	}
	var diagnostics []string
	for line := range lines {
		diagnostics = append(diagnostics, line)
	}
	if strings.Join(diagnostics, "\n") != "context canceled\ninterrupt signal received" {
		t.Fatalf("stderr lines=%q", diagnostics)
	}
}
