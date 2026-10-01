//go:build darwin || linux

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The limit applies only to the child. Its first two-byte record succeeds,
// then the four-byte replacement encounters a partial write and EFBIG.
func TestCLIPartialWriteRetainsAcceptedDuplicate(t *testing.T) {
	binary := buildCLI(t)
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestFileSizeLimitHelper$")
	cmd.Env = append(os.Environ(), "LEDGERLOAD_LIMIT_HELPER=1", "LEDGERLOAD_BINARY="+binary, "LEDGERLOAD_DIR="+dir)
	cmd.Stdin = strings.NewReader("keep,1\nkeep,123\nlater,2\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("limited executable exceeded deadline: %v", ctx.Err())
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("limited executable = %v; stderr %q", err, stderr.String())
	}
	assertProcess(t, exitErr.ExitCode(), stdout.String(), stderr.String(), 2, "file too large")
	assertRecords(t, dir, map[string]string{"keep": "1\n"})
}

func TestFileSizeLimitHelper(t *testing.T) {
	if os.Getenv("LEDGERLOAD_LIMIT_HELPER") != "1" {
		return
	}
	signal.Ignore(syscall.SIGXFSZ)
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 2, Max: 2}); err != nil {
		fmt.Fprintln(os.Stderr, "set child file-size limit:", err)
		os.Exit(99)
	}
	binary := os.Getenv("LEDGERLOAD_BINARY")
	args := []string{binary, "--dir", os.Getenv("LEDGERLOAD_DIR")}
	if err := syscall.Exec(filepath.Clean(binary), args, os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, "exec ledgerload:", err)
		os.Exit(99)
	}
}
