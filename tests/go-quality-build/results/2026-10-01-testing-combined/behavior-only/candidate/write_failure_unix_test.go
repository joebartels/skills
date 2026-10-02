//go:build darwin || linux || freebsd

package indexer_test

import (
	"context"
	"errors"
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

func TestPartialWritesPreservePriorState(t *testing.T) {
	for _, operation := range []string{"put", "apply", "refresh"} {
		t.Run(operation, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestWriteFailureChild$", "-test.timeout=4s")
			command.Env = append(os.Environ(), "INDEXER_WRITE_FAILURE_CHILD="+operation)
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("limited child: %v\n%s", err, output)
			}
		})
	}
}

func TestWriteFailureChild(t *testing.T) {
	operation := os.Getenv("INDEXER_WRITE_FAILURE_CHILD")
	if operation == "" {
		t.Skip("only runs in a file-size-limited child")
	}
	root := t.TempDir()
	i := indexer.Open(root)
	old := strings.Repeat("original ", 1024)
	if err := i.Put("alpha", old); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "records.json")
	if err := os.WriteFile(path, []byte(old), 0600); err != nil {
		t.Fatal(err)
	}
	// A 128 KiB publication must fail after its initial 4 KiB write. Limits and
	// SIGXFSZ handling belong only to this disposable child, never the test host.
	signal.Ignore(syscall.SIGXFSZ)
	limit := syscall.Rlimit{Cur: 4096, Max: 4096}
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &limit); err != nil {
		t.Fatal(err)
	}
	large := strings.Repeat("x", 128*1024)
	var err error
	switch operation {
	case "put":
		err = i.Put("alpha", large)
		assertRecord(t, i, "alpha", old)
	case "apply":
		err = i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,accepted too\nalpha," + large + "\ngamma,later\n"))
		assertRecord(t, i, "alpha", "accepted")
		assertRecord(t, i, "beta", "accepted too")
		assertMissing(t, i, "gamma")
	case "refresh":
		body := &trackedBody{Reader: strings.NewReader(`[{"key":"alpha","text":"` + large + `"}]`)}
		err = indexer.Refresh(context.Background(), bodyClient(body, 200), "https://example.test/records", path)
		assertFile(t, path, old)
		if body.closes != 1 {
			t.Fatalf("body closes = %d", body.closes)
		}
	default:
		t.Fatalf("unknown operation %q", operation)
	}
	if !errors.Is(err, syscall.EFBIG) {
		t.Fatalf("error = %v, want actual file-size write failure", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".record-") || strings.HasPrefix(entry.Name(), ".refresh-") {
			t.Errorf("temporary file leaked: %s", entry.Name())
		}
	}
}
