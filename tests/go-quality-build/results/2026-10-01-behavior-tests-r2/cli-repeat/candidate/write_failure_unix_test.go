//go:build darwin || linux

package main

import (
	"context"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// This helper changes only its child process's limit before executing ledgerload.
func TestFileSizeLimitHelper(t *testing.T) {
	if os.Getenv("LEDGERLOAD_LIMIT_HELPER") != "1" {
		return
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 2, Max: 2}); err != nil {
		t.Fatal(err)
	}
	binary := os.Getenv("LEDGERLOAD_TEST_BINARY")
	if err := syscall.Exec(binary, []string{binary, "--dir", os.Getenv("LEDGERLOAD_TEST_DIR")}, os.Environ()); err != nil {
		t.Fatal(err)
	}
}

func TestPartialWriteRetainsAcceptedReplacement(t *testing.T) {
	binary := buildCLI(t)
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestFileSizeLimitHelper$")
	cmd.Env = append(os.Environ(), "LEDGERLOAD_LIMIT_HELPER=1", "LEDGERLOAD_TEST_BINARY="+binary, "LEDGERLOAD_TEST_DIR="+dir)
	checkProcess(t, cmd, "web,7\nweb,12345\nlater,4\n", 2)
	assertFiles(t, dir, map[string]string{"web": "7\n"})
}
