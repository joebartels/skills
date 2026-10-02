//go:build darwin || linux

package indexer_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"example.com/indexer"
)

// File-size limits are confined to a child process. They force partial progress
// while writing, rather than merely preventing a temporary file from opening.
func TestPartialWriteRetainsPriorBytes(t *testing.T) {
	for _, operation := range []string{"put", "csv", "refresh"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			path := filepath.Join(root, "alpha")
			prior := "exact old bytes\x00\n"
			if err := os.WriteFile(path, []byte(prior), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPartialWriteChild$", "-test.timeout=4s")
			cmd.Env = []string{"INDEXER_WRITE_CHILD=" + operation, "INDEXER_WRITE_ROOT=" + root}
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("write-failure child: %v\n%s", err, output)
			}
			t.Logf("bounded child output: %s", output)
			if !strings.Contains(string(output), "partial write rejected") {
				t.Fatalf("child did not exercise expected failure: %s", output)
			}
			if operation == "csv" {
				assertFile(t, path, "accepted replacement")
				assertFile(t, filepath.Join(root, "beta"), "accepted prefix")
				if _, err := os.Stat(filepath.Join(root, "gamma")); !os.IsNotExist(err) {
					t.Errorf("later CSV row exists: %v", err)
				}
			} else {
				assertFile(t, path, prior)
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".record-") {
					t.Errorf("failed temporary file remains: %s", entry.Name())
				}
			}
		})
	}
}

func TestPartialWriteChild(t *testing.T) {
	operation := os.Getenv("INDEXER_WRITE_CHILD")
	if operation == "" {
		t.Skip("subprocess-only file limit fixture")
	}
	root := os.Getenv("INDEXER_WRITE_ROOT")
	signal.Ignore(syscall.SIGXFSZ)
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 1024, Max: 1024}); err != nil {
		t.Fatal(err)
	}
	large := strings.Repeat("x", 64*1024)
	var err error
	switch operation {
	case "put":
		err = indexer.Open(root).Put("alpha", large)
	case "csv":
		err = indexer.Open(root).ApplyCSV(strings.NewReader("beta,accepted prefix\nalpha,accepted replacement\nalpha," + large + "\ngamma,later\n"))
	case "refresh":
		client := &http.Client{Transport: transportFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: &observedBody{Reader: strings.NewReader(`[{"key":"alpha","text":"` + large + `"}]`)}, Header: make(http.Header), Request: req}, nil
		})}
		err = indexer.Refresh(context.Background(), client, "http://fixture.invalid", filepath.Join(root, "alpha"))
	default:
		t.Fatalf("unknown child operation %q", operation)
	}
	if err == nil {
		t.Fatal("oversized file unexpectedly succeeded")
	}
	if !errors.Is(err, syscall.EFBIG) && !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("expected write-stage limit error, got %v", err)
	}
	fmt.Println("partial write rejected:", err)
}
