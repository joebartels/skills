//go:build unix

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

func TestAtomicRetentionAfterPartialWrite(t *testing.T) {
	for _, mode := range []string{"put", "apply", "refresh"} {
		t.Run(mode, func(t *testing.T) {
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "-test.run=^TestAtomicWriteLimitHelper$", "-test.timeout=4s")
			cmd.Env = append(os.Environ(), "INDEXER_TEST_WRITE_LIMIT="+mode, "INDEXER_TEST_ROOT="+t.TempDir())
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("partial-write child (%s): %v\n%s", mode, err, output)
			}
		})
	}
}

func TestAtomicWriteLimitHelper(t *testing.T) {
	mode := os.Getenv("INDEXER_TEST_WRITE_LIMIT")
	if mode == "" {
		return
	}
	root := os.Getenv("INDEXER_TEST_ROOT")
	path := filepath.Join(root, "alpha")
	if err := os.WriteFile(path, []byte("prior bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	// The limit is local to this child. Ignoring SIGXFSZ lets File.Write
	// return its partial-write error instead of terminating the child.
	signal.Ignore(syscall.SIGXFSZ)
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 1024, Max: 1024}); err != nil {
		t.Fatal(err)
	}
	i := indexer.Open(root)
	large := strings.Repeat("x", 8192)
	var err error
	want := "prior bytes"
	switch mode {
	case "put":
		err = i.Put("alpha", large)
	case "apply":
		err = i.ApplyCSV(strings.NewReader("alpha,accepted\nalpha," + large + "\nlater,forbidden\n"))
		want = "accepted"
	case "refresh":
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`[{"key":"alpha","text":"` + large + `"}]`)), Header: make(http.Header)}, nil
		})}
		err = indexer.Refresh(context.Background(), client, "http://example.test/records", path)
	default:
		t.Fatalf("unknown helper mode %q", mode)
	}
	if err == nil {
		t.Fatal("large write succeeded under finite file-size limit")
	}
	if !errors.Is(err, syscall.EFBIG) && !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("error = %v; want failure from the limited write", err)
	}
	assertFile(t, path, want)
	assertMissing(t, i, "later")
	entries, readErr := os.ReadDir(root)
	if readErr != nil || len(entries) != 1 || entries[0].Name() != "alpha" {
		t.Fatalf("files after failed write = %v, %v", entries, readErr)
	}
}
