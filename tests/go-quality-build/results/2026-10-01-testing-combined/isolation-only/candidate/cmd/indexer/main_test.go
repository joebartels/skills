package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestExistingPut(t *testing.T) {
	if err := run([]string{"put", t.TempDir(), "alpha", "one"}); err != nil {
		t.Fatal(err)
	}
	if err := run(nil); err == nil {
		t.Fatal("missing usage rejection")
	}
}

func TestCommandContracts(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "indexer")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	build.Env = childEnvironment()
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build command: %v\n%s", err, output)
	}

	t.Run("put-success", func(t *testing.T) {
		t.Parallel()
		cwd := t.TempDir()
		// Positional arguments beginning with '-' retain the existing grammar.
		result := runProcess(t, binary, cwd, "", "put", "-records", "alpha", "-literal\n")
		assertOutcome(t, result, 0)
		assertFile(t, filepath.Join(cwd, "-records", "alpha"), "-literal\n")
	})
	t.Run("put-empty-text", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		result := runProcess(t, binary, root, "", "put", root, "alpha", "")
		assertOutcome(t, result, 0)
		assertFile(t, filepath.Join(root, "alpha"), "")
	})
	t.Run("put-rejection", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		result := runProcess(t, binary, root, "", "put", root, "../escape", "bad")
		assertOutcome(t, result, 2)
		if _, err := os.Stat(filepath.Join(root, "escape")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("unexpected output file: %v", err)
		}
	})
	t.Run("apply-success", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		result := runProcess(t, binary, root, "alpha,old\nbeta,\"comma, newline\nkept\"\nalpha,new\n", "apply", "--dir", root)
		assertOutcome(t, result, 0)
		assertFile(t, filepath.Join(root, "alpha"), "new")
		assertFile(t, filepath.Join(root, "beta"), "comma, newline\nkept")
	})
	t.Run("apply-rejected-prefix", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		result := runProcess(t, binary, root, "alpha,accepted\nalpha, \nbeta,later\n", "apply", "--dir", root)
		assertOutcome(t, result, 2)
		assertFile(t, filepath.Join(root, "alpha"), "accepted")
		if _, err := os.Stat(filepath.Join(root, "beta")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("later row exists: %v", err)
		}
	})
	t.Run("apply-parse-prefix", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		result := runProcess(t, binary, root, "alpha,accepted\nbeta,\"bad\n", "apply", "--dir", root)
		assertOutcome(t, result, 2)
		assertFile(t, filepath.Join(root, "alpha"), "accepted")
	})
	t.Run("apply-write-prefix", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "beta"), 0700); err != nil {
			t.Fatal(err)
		}
		result := runProcess(t, binary, root, "alpha,accepted\nbeta,rejected\ngamma,later\n", "apply", "--dir", root)
		assertOutcome(t, result, 2)
		assertFile(t, filepath.Join(root, "alpha"), "accepted")
		if _, err := os.Stat(filepath.Join(root, "gamma")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("later row exists: %v", err)
		}
	})
	invalidArgs := []struct {
		name string
		args []string
	}{
		{"missing", nil}, {"unknown", []string{"unknown"}},
		{"put-arity", []string{"put", "dir", "key"}},
		{"apply-missing-dir", []string{"apply"}},
		{"apply-unknown-flag", []string{"apply", "--unknown"}},
		{"apply-extra", []string{"apply", "--dir", "dir", "extra"}},
		{"watch-missing", []string{"watch"}},
		{"watch-scheme", []string{"watch", "--url", "file:///tmp/records", "--file", "out", "--interval", "1s"}},
		{"watch-host", []string{"watch", "--url", "http:///records", "--file", "out", "--interval", "1s"}},
		{"watch-url-parse", []string{"watch", "--url", "http://%", "--file", "out", "--interval", "1s"}},
		{"watch-path", []string{"watch", "--url", "http://localhost", "--file", "", "--interval", "1s"}},
		{"watch-zero", []string{"watch", "--url", "http://localhost", "--file", "out", "--interval", "0s"}},
		{"watch-negative", []string{"watch", "--url", "http://localhost", "--file", "out", "--interval", "-1s"}},
		{"watch-duration", []string{"watch", "--url", "http://localhost", "--file", "out", "--interval", "bad"}},
		{"watch-extra", []string{"watch", "--url", "http://localhost", "--file", "out", "--interval", "1s", "extra"}},
	}
	for _, test := range invalidArgs {
		t.Run("startup/"+test.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			assertOutcome(t, runProcess(t, binary, root, "", test.args...), 2)
			entries, err := os.ReadDir(root)
			if err != nil || len(entries) != 0 {
				t.Fatalf("startup wrote files: %v, %v", entries, err)
			}
		})
	}
	t.Run("watch-first-refresh-failure", func(t *testing.T) {
		t.Parallel()
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1); w.WriteHeader(503) }))
		t.Cleanup(server.Close)
		root := t.TempDir()
		path := filepath.Join(root, "snapshot.json")
		if err := os.WriteFile(path, []byte("exact prior bytes"), 0600); err != nil {
			t.Fatal(err)
		}
		assertOutcome(t, runProcess(t, binary, root, "", "watch", "--url", server.URL, "--file", path, "--interval", "20ms"), 2)
		assertFile(t, path, "exact prior bytes")
		if requests.Load() != 1 {
			t.Fatalf("requests=%d; want exactly 1", requests.Load())
		}
	})
	for _, signal := range []struct {
		name  string
		value os.Signal
	}{{"interrupt", os.Interrupt}, {"termination", syscall.SIGTERM}} {
		t.Run("watch-recurrence-and-"+signal.name, func(t *testing.T) {
			t.Parallel()
			if runtime.GOOS == "windows" {
				t.Skip("POSIX process signals are the exercised contract")
			}
			requests := make(chan int, 8)
			secondGate := make(chan struct{})
			shutdownGate := make(chan struct{})
			handlerCanceled := make(chan struct{})
			var releaseSecond, releaseShutdown sync.Once
			var count atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := int(count.Add(1))
				requests <- n
				if n == 2 {
					select {
					case <-secondGate:
					case <-r.Context().Done():
						return
					}
				}
				if n == 3 {
					select {
					case <-r.Context().Done():
						close(handlerCanceled)
						return
					case <-shutdownGate:
						return
					}
				}
				fmt.Fprintf(w, `[{"key":"alpha","text":%q}]`, strconv.Itoa(n))
			}))
			t.Cleanup(server.Close)
			root := t.TempDir()
			path := filepath.Join(root, "snapshot.json")
			processCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			cmd := exec.CommandContext(processCtx, binary, "watch", "--url", server.URL, "--file", path, "--interval", "25ms")
			cmd.Dir, cmd.Env = root, childEnvironment()
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Start(); err != nil {
				cancel()
				t.Fatal(err)
			}
			waitResult := make(chan error, 1)
			finished := make(chan struct{})
			t.Cleanup(func() {
				releaseSecond.Do(func() { close(secondGate) })
				releaseShutdown.Do(func() { close(shutdownGate) })
				cancel()
				cmd.Process.Kill()
				select {
				case <-finished:
				case <-time.After(3 * time.Second):
					t.Error("watch process did not join during cleanup")
				}
			})
			go func() { defer close(finished); waitResult <- cmd.Wait() }()
			awaitRequest(t, requests, waitResult, 1)
			awaitRequest(t, requests, waitResult, 2)
			assertFile(t, path, "[{\"key\":\"alpha\",\"text\":\"1\"}]\n")
			releaseSecond.Do(func() { close(secondGate) })
			awaitRequest(t, requests, waitResult, 3)
			assertFile(t, path, "[{\"key\":\"alpha\",\"text\":\"2\"}]\n")
			if err := cmd.Process.Signal(signal.value); err != nil {
				t.Fatal(err)
			}
			select {
			case <-handlerCanceled:
			case <-time.After(3 * time.Second):
				t.Fatal("signal did not cancel in-flight HTTP request")
			}
			select {
			case err := <-waitResult:
				if err != nil {
					t.Fatalf("watch exit: %v; stderr=%s", err, stderr.String())
				}
			case <-time.After(3 * time.Second):
				t.Fatal("watch did not join after signal")
			}
			select {
			case <-finished:
			case <-time.After(3 * time.Second):
				t.Fatal("watch worker did not join")
			}
			if stdout.Len() != 0 || stderr.Len() != 0 {
				t.Fatalf("successful watch stdout=%q stderr=%q; want empty", stdout.String(), stderr.String())
			}
			assertFile(t, path, "[{\"key\":\"alpha\",\"text\":\"2\"}]\n")
		})
	}
}

type processResult struct {
	stdout, stderr string
	code           int
}

func runProcess(t *testing.T, binary, cwd, input string, args ...string) processResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir, cmd.Env, cmd.Stdin = cwd, childEnvironment(), strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("process %v exceeded deadline: %v; stderr=%s", args, ctx.Err(), stderr.String())
	}
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("run process: %v", err)
		}
		code = exitErr.ExitCode()
	}
	return processResult{stdout.String(), stderr.String(), code}
}
func assertOutcome(t *testing.T, result processResult, code int) {
	t.Helper()
	if result.code != code || result.stdout != "" || (code == 0 && result.stderr != "") || (code != 0 && result.stderr == "") {
		t.Fatalf("process = exit %d stdout=%q stderr=%q; want exit %d, no stdout, diagnostic only on failure", result.code, result.stdout, result.stderr, code)
	}
}
func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("ReadFile(%q) = %q, %v; want %q", path, got, err, want)
	}
}
func childEnvironment() []string {
	// Explicit allowlist avoids inheriting unrelated runtime/configuration knobs.
	env := []string{"GOTOOLCHAIN=local", "GOCACHE=/private/tmp/go-quality-testing-cache"}
	for _, name := range []string{"PATH", "HOME", "TMPDIR", "SYSTEMROOT"} {
		if value, exists := os.LookupEnv(name); exists {
			env = append(env, name+"="+value)
		}
	}
	return env
}
func awaitRequest(t *testing.T, requests <-chan int, process <-chan error, want int) {
	t.Helper()
	select {
	case got := <-requests:
		if got != want {
			t.Fatalf("request=%d; want %d", got, want)
		}
	case err := <-process:
		t.Fatalf("watch exited before request %d: %v", want, err)
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for watch request %d", want)
	}
}
