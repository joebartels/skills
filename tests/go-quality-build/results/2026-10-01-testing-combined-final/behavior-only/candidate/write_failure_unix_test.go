//go:build darwin || linux

package indexer_test

import (
	"context"
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

// A child-local file-size limit produces a real write failure after progress.
// This catches truncating the destination before a failing write, unlike an
// obstruction that only prevents opening or renaming a file.
func TestWriteFailureAfterProgressRetainsState(t *testing.T) {
	for _, mode := range []string{"put", "csv", "refresh"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "alpha")
			if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run=^TestWriteFailureHelper$")
			cmd.Env = append(os.Environ(), "INDEXER_WRITE_FAILURE="+mode, "INDEXER_WRITE_ROOT="+root)
			output, err := cmd.CombinedOutput()
			if ctx.Err() != nil || err != nil || !strings.Contains(string(output), "write rejected") {
				t.Fatalf("helper = %v (context %v), output %s", err, ctx.Err(), output)
			}
			want := "original"
			if mode == "csv" {
				want = "accepted"
				assertFile(t, filepath.Join(root, "beta"), "prefix")
			}
			assertFile(t, path, want)
			if _, err := os.Stat(filepath.Join(root, "gamma")); !os.IsNotExist(err) {
				t.Fatalf("later effect: %v", err)
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".record-") {
					t.Fatalf("temporary file leaked: %s", entry.Name())
				}
			}
		})
	}
}

func TestWriteFailureHelper(t *testing.T) {
	mode := os.Getenv("INDEXER_WRITE_FAILURE")
	if mode == "" {
		return
	}
	signal.Ignore(syscall.SIGXFSZ)
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 64, Max: 64}); err != nil {
		t.Fatal(err)
	}
	root := os.Getenv("INDEXER_WRITE_ROOT")
	text := strings.Repeat("x", 1024)
	var err error
	switch mode {
	case "put":
		err = indexer.Open(root).Put("alpha", text)
	case "csv":
		err = indexer.Open(root).ApplyCSV(strings.NewReader("alpha,accepted\nbeta,prefix\nalpha," + text + "\ngamma,later\n"))
	case "refresh":
		body := io.NopCloser(strings.NewReader(`[{"key":"alpha","text":"` + text + `"}]`))
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return &http.Response{StatusCode: 200, Body: body}, nil })}
		err = indexer.Refresh(context.Background(), client, "https://records.test", filepath.Join(root, "alpha"))
	default:
		t.Fatalf("unknown mode %q", mode)
	}
	if err == nil {
		t.Fatal("write unexpectedly succeeded")
	}
	fmt.Println("write rejected:", err)
}
