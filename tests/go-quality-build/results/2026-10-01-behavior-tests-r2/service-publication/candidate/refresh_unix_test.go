//go:build darwin || linux

package mirror_test

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

	"example.com/mirror"
)

func TestRefreshReplacesSnapshotWithoutTruncatingOpenHandle(t *testing.T) {
	const old = "old snapshot with exact spaces and a newline\n"
	path := filepath.Join(t.TempDir(), "items.json")
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	oldHandle, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer oldHandle.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, tc := range []struct{ input, want string }{
		{`[{"code":"new","label":"First replacement"}]`, `[{"code":"new","label":"First replacement"}]`},
		{`[]`, `[]`},
	} {
		body := &trackedBody{Reader: strings.NewReader(tc.input)}
		if err := mirror.Refresh(ctx, responseClient(http.StatusOK, body), "https://mirror.invalid/items", path); err != nil {
			t.Fatal(err)
		}
		requireFileBytes(t, path, tc.want)
		if _, err := oldHandle.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}
		got, err := io.ReadAll(oldHandle)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != old {
			t.Fatalf("already-open old handle = %q, want %q", got, old)
		}
	}
	// A later rejected refresh must retain the last accepted replacement.
	body := &trackedBody{Reader: strings.NewReader(`[{"code":"","label":"Invalid"}]`)}
	if err := mirror.Refresh(ctx, responseClient(http.StatusOK, body), "https://mirror.invalid/items", path); err == nil {
		t.Fatal("Refresh accepted invalid replacement")
	}
	requireFileBytes(t, path, `[]`)
	requireOnlySnapshot(t, path)
}

func TestRefreshPartialWriteRetainsPriorSnapshot(t *testing.T) {
	const old = "previous exact snapshot bytes, longer than the child file-size limit, must all survive\n"
	path := filepath.Join(t.TempDir(), "items.json")
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestRefreshPartialWriteHelper$", "-test.timeout=5s")
	cmd.Env = append(os.Environ(), "MIRROR_PARTIAL_WRITE_PATH="+path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("partial-write subprocess: %v\n%s", err, output)
	}
	requireFileBytes(t, path, old)
	requireOnlySnapshot(t, path)
}

func TestRefreshPartialWriteHelper(t *testing.T) {
	path := os.Getenv("MIRROR_PARTIAL_WRITE_PATH")
	if path == "" {
		t.Skip("runs only as the bounded partial-write subprocess")
	}
	// Limit only this process; SIGXFSZ must produce a write error rather than
	// terminating the child before the library can clean up its staged file.
	signal.Ignore(syscall.SIGXFSZ)
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 64, Max: 64}); err != nil {
		t.Fatal(err)
	}
	body := &trackedBody{Reader: strings.NewReader(`[{"code":"new","label":"` + strings.Repeat("snapshot", 128) + `"}]`)}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := mirror.Refresh(ctx, responseClient(http.StatusOK, body), "https://mirror.invalid/items", path); !errors.Is(err, syscall.EFBIG) {
		t.Fatalf("Refresh error = %v, want the partial-write file-size error", err)
	}
	if body.closes != 1 {
		t.Errorf("body close calls = %d, want 1", body.closes)
	}
}
