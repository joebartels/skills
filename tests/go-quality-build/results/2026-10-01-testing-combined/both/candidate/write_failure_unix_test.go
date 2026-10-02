//go:build darwin || linux

package indexer_test

import (
	"context"
	"errors"
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

func TestPartialWriteFailureRetainsPriorData(t *testing.T) {
	for _, mode := range []string{"put", "csv", "refresh"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, filepath.Join(root, "alpha"), "exact prior bytes\x00\n")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestFileSizeLimitHelper$", "-test.timeout=5s")
			cmd.Env = []string{"GOCACHE=/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN=local", "GO_INDEXER_LIMIT_MODE=" + mode, "GO_INDEXER_LIMIT_ROOT=" + root}
			for _, key := range []string{"PATH", "HOME", "TMPDIR"} {
				if value, ok := os.LookupEnv(key); ok {
					cmd.Env = append(cmd.Env, key+"="+value)
				}
			}
			cmd.WaitDelay = time.Second
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("file-size-limited child: %v\n%s", err, output)
			}
			want := "exact prior bytes\x00\n"
			if mode == "csv" {
				want = "accepted"
				assertFile(t, filepath.Join(root, "beta"), "accepted beta")
			}
			assertFile(t, filepath.Join(root, "alpha"), want)
			if _, err := os.Stat(filepath.Join(root, "gamma")); !os.IsNotExist(err) {
				t.Fatalf("later row was written: %v", err)
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".record-") {
					t.Errorf("failed write leaked %s", entry.Name())
				}
			}
		})
	}
}

// This helper runs only in an owned child. The finite RLIMIT exercises a real
// write failure after bytes have been written, without limiting the test host.
func TestFileSizeLimitHelper(t *testing.T) {
	mode := os.Getenv("GO_INDEXER_LIMIT_MODE")
	if mode == "" {
		return
	}
	root := os.Getenv("GO_INDEXER_LIMIT_ROOT")
	signal.Ignore(syscall.SIGXFSZ)
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 1024, Max: 1024}); err != nil {
		t.Fatalf("set child file-size limit: %v", err)
	}
	huge := strings.Repeat("x", 8192)
	var err error
	switch mode {
	case "put":
		err = indexer.Open(root).Put("alpha", huge)
	case "csv":
		err = indexer.Open(root).ApplyCSV(strings.NewReader("alpha,accepted\nbeta,accepted beta\nalpha," + huge + "\ngamma,later\n"))
	case "refresh":
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("[{\"key\":\"alpha\",\"text\":\"" + huge + "\"}]")), Header: make(http.Header)}, nil
		})}
		err = indexer.Refresh(context.Background(), client, "http://records.example", filepath.Join(root, "alpha"))
	default:
		t.Fatalf("unknown child mode %q", mode)
	}
	if !errors.Is(err, syscall.EFBIG) && !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("write failure = %v; want file-size or short-write error", err)
	}
}
