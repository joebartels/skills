package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestCommandProcesses(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "indexer")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build = %v, %s", err, output)
	}

	t.Run("put contract", func(t *testing.T) {
		root := t.TempDir()
		// Preserve the old positional grammar, including option-shaped text.
		command(t, binary, []string{"put", root, "alpha", "--literal\n"}, "", 0)
		checkFile(t, filepath.Join(root, "alpha"), "--literal\n")
		command(t, binary, []string{"put", root, "alpha", ""}, "", 0)
		checkFile(t, filepath.Join(root, "alpha"), "")
		command(t, binary, []string{"put", root, "Alpha", "bad"}, "", 2)
		checkFile(t, filepath.Join(root, "alpha"), "")
		command(t, binary, nil, "", 2)
		command(t, binary, []string{"put", root, "alpha"}, "", 2)
	})
	t.Run("apply success", func(t *testing.T) {
		root := t.TempDir()
		command(t, binary, []string{"apply", "--dir", root}, "alpha,one\nbeta,\" two, three \"\nalpha,last\n", 0)
		checkFile(t, filepath.Join(root, "alpha"), "last")
		checkFile(t, filepath.Join(root, "beta"), " two, three ")
	})
	for _, row := range []string{"alpha,\" \t\"\n", "alpha,replacement,extra\n"} {
		t.Run("apply rejected prefix "+row, func(t *testing.T) {
			root := t.TempDir()
			command(t, binary, []string{"apply", "--dir", root}, "alpha,accepted\nbeta,prefix\n"+row+"gamma,later\n", 2)
			checkFile(t, filepath.Join(root, "alpha"), "accepted")
			checkFile(t, filepath.Join(root, "beta"), "prefix")
			if _, err := os.Stat(filepath.Join(root, "gamma")); !os.IsNotExist(err) {
				t.Fatalf("later row exists: %v", err)
			}
		})
	}
	t.Run("apply usage", func(t *testing.T) { command(t, binary, []string{"apply"}, "", 2) })
	t.Run("watch startup rejection", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "snapshot")
		for _, args := range [][]string{
			{"watch"},
			{"watch", "--url", "file:///tmp/data", "--file", path, "--interval", "1s"},
			{"watch", "--url", "http:///missing-host", "--file", path, "--interval", "1s"},
			{"watch", "--url", "https://records.test", "--file", "", "--interval", "1s"},
			{"watch", "--url", "https://records.test", "--file", path, "--interval", "0"},
			{"watch", "--url", "https://records.test", "--file", path, "--interval", "-1s"},
			{"watch", "--url", "https://records.test", "--file", path, "--interval", "oops"},
		} {
			command(t, binary, args, "", 2)
		}
	})
	t.Run("watch failed first refresh", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			w.WriteHeader(206)
			fmt.Fprint(w, `[{"key":"alpha","text":"valid"}]`)
		}))
		defer server.Close()
		path := filepath.Join(t.TempDir(), "snapshot")
		if err := os.WriteFile(path, []byte("prior\x00\n"), 0600); err != nil {
			t.Fatal(err)
		}
		command(t, binary, []string{"watch", "--url", server.URL, "--file", path, "--interval", "5ms"}, "", 2)
		checkFile(t, path, "prior\x00\n")
		if calls.Load() != 1 {
			t.Fatalf("GET calls = %d", calls.Load())
		}
	})
	for _, signal := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		t.Run("watch recurrence and "+signal.String(), func(t *testing.T) { watchProcess(t, binary, signal) })
	}
}

func command(t *testing.T, binary string, args []string, input string, wantCode int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("command deadline: %v", ctx.Err())
	}
	code := 0
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	if code != wantCode || stdout.Len() != 0 || (wantCode == 0 && stderr.Len() != 0) || (wantCode != 0 && stderr.Len() == 0) {
		t.Fatalf("%v: exit %d (want %d), stdout %q, stderr %q", args, code, wantCode, stdout.String(), stderr.String())
	}
}

func watchProcess(t *testing.T, binary string, signal os.Signal) {
	var calls atomic.Int32
	third, canceled := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if r.Method != http.MethodGet {
			t.Errorf("method = %s", r.Method)
		}
		if n == 3 {
			close(third)
			<-r.Context().Done()
			close(canceled)
			return
		}
		fmt.Fprintf(w, `[{"key":"alpha","text":"cycle %d"}]`, n)
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "snapshot")
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "watch", "--url", server.URL, "--file", path, "--interval", "15ms")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	result, done := make(chan error, 1), make(chan struct{})
	go func() { result <- cmd.Wait(); close(done) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("child did not join after cleanup")
		}
	})
	select {
	case <-third:
	case <-done:
		t.Fatalf("watch stopped before recurrence: %v, stdout %q, stderr %q", <-result, stdout.String(), stderr.String())
	case <-time.After(3 * time.Second):
		cancel()
		<-done
		t.Fatalf("watch did not recur: stdout %q, stderr %q", stdout.String(), stderr.String())
	}
	// Starting the third request proves both earlier publications completed.
	checkFile(t, path, "[{\"key\":\"alpha\",\"text\":\"cycle 2\"}]\n")
	if err := cmd.Process.Signal(signal); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("watch did not stop after signal")
	}
	if err := <-result; err != nil {
		t.Fatalf("watch = %v, stderr %q", err, stderr.String())
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("in-flight request was not canceled")
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("stdout %q, stderr %q", stdout.String(), stderr.String())
	}
	checkFile(t, path, "[{\"key\":\"alpha\",\"text\":\"cycle 2\"}]\n")
	if calls.Load() != 3 {
		t.Fatalf("requests after cancellation: %d", calls.Load())
	}
}

func checkFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file = %q, %v; want %q", got, err, want)
	}
}
