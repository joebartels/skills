//go:build linux || darwin

package mirror_test

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

	mirror "example.com/mirror"
)

func TestRefreshReplacesSnapshotWithoutTruncatingOldHandle(t *testing.T) {
	previous := "old snapshot with exact bytes\n"
	path := snapshotPath(t, previous)
	old, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer old.Close()
	body := &observedBody{Reader: strings.NewReader(`[{"code":"new","label":"New"}]`)}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
	})}
	if err := mirror.Refresh(context.Background(), client, "http://snapshot.invalid/items", path); err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(old)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != previous {
		t.Fatalf("already-open snapshot = %q, want %q", got, previous)
	}
	assertSnapshot(t, path, `[{"code":"new","label":"New"}]`)
	assertOnlySnapshot(t, path)
}

func TestRefreshPartialWriteFailureRetainsPreviousSnapshot(t *testing.T) {
	if path := os.Getenv("MIRROR_PARTIAL_WRITE_CHILD_PATH"); path != "" {
		signal.Ignore(syscall.SIGXFSZ)
		var limit syscall.Rlimit
		if err := syscall.Getrlimit(syscall.RLIMIT_FSIZE, &limit); err != nil {
			t.Fatal(err)
		}
		limit.Cur = 64
		if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &limit); err != nil {
			t.Fatal(err)
		}
		// Confirm this process's filesystem boundary permits some progress and
		// then fails. An open obstruction alone would not test partial writes.
		probe, err := os.Create(filepath.Join(filepath.Dir(path), "limit-probe"))
		if err != nil {
			t.Fatal(err)
		}
		n, writeErr := probe.Write(make([]byte, 128))
		closeErr := probe.Close()
		if err := os.Remove(probe.Name()); err != nil {
			t.Fatal(err)
		}
		if n != 64 || writeErr == nil || closeErr != nil {
			t.Fatalf("partial-write probe = %d bytes, write error %v, close error %v", n, writeErr, closeErr)
		}
		body := &observedBody{Reader: strings.NewReader(`[{"code":"new","label":"` + strings.Repeat("x", 4096) + `"}]`)}
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
		})}
		if err := mirror.Refresh(context.Background(), client, "http://snapshot.invalid/items", path); err == nil {
			t.Fatal("Refresh succeeded despite the partial-write limit")
		}
		if body.closes != 1 {
			t.Fatalf("response body close calls = %d, want 1", body.closes)
		}
		fmt.Println("partial-write boundary exercised")
		return
	}
	previous := "accepted snapshot\n"
	path := snapshotPath(t, previous)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestRefreshPartialWriteFailureRetainsPreviousSnapshot$", "-test.timeout=8s")
	cmd.Env = append(os.Environ(), "MIRROR_PARTIAL_WRITE_CHILD_PATH="+path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("partial-write subprocess failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "partial-write boundary exercised") {
		t.Fatalf("partial-write subprocess did not exercise the limit:\n%s", output)
	}
	assertSnapshot(t, path, previous)
	assertOnlySnapshot(t, path)
}
