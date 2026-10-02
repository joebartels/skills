//go:build darwin || linux

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"testing"
	"time"
)

func TestCLIWriteFailure(t *testing.T) {
	binary := buildCLI(t)
	for _, tc := range []struct {
		name       string
		limit      uint64
		input      string
		want       map[string]string
		diagnostic string
	}{
		{"first write", 0, "web,7\nlater,2\n", nil, "record 1"},
		{"partial duplicate write", 4, "web,7\nweb,123456\nlater,2\n", map[string]string{"web": "7\n"}, "record 2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			// Change only a short-lived child's per-file limit, then exec the real CLI.
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestFileSizeLimitHelper$", "--", binary, dir, strconv.FormatUint(tc.limit, 10))
			cmd.Env = append(os.Environ(), "LEDGERLOAD_FILE_SIZE_LIMIT_HELPER=1")
			checkProcess(t, cmd, tc.input, 2, tc.diagnostic)
			checkFiles(t, dir, tc.want)
		})
	}
}

func TestFileSizeLimitHelper(t *testing.T) {
	if os.Getenv("LEDGERLOAD_FILE_SIZE_LIMIT_HELPER") != "1" {
		return
	}
	args := os.Args[len(os.Args)-3:]
	limit, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse file-size limit:", err)
		os.Exit(125)
	}
	// Make an over-limit write return EFBIG instead of terminating the process.
	signal.Ignore(syscall.SIGXFSZ)
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: limit, Max: limit}); err != nil {
		fmt.Fprintln(os.Stderr, "set child file-size limit:", err)
		os.Exit(125)
	}
	if err := syscall.Exec(args[0], []string{args[0], "--dir", args[1]}, os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, "exec CLI:", err)
		os.Exit(125)
	}
}
