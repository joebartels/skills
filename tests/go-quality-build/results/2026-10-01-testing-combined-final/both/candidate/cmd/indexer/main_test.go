package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

var executable string

func commandEnvironment() []string {
	return []string{"PATH=" + os.Getenv("PATH"), "GOCACHE=" + os.Getenv("GOCACHE"), "GOTOOLCHAIN=local", "GOENV=off", "TMPDIR=" + os.TempDir()}
}

func TestMain(m *testing.M) {
	root, err := os.MkdirTemp("", "indexer-command-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	executable = filepath.Join(root, "indexer")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	build := exec.CommandContext(ctx, "go", "build", "-o", executable, ".")
	build.Env = commandEnvironment()
	output, err := build.CombinedOutput()
	cancel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build executable: %v\n%s", err, output)
		os.RemoveAll(root)
		os.Exit(1)
	}
	code := m.Run()
	if err := os.RemoveAll(root); err != nil {
		fmt.Fprintln(os.Stderr, "remove command fixture:", err)
		code = 1
	}
	os.Exit(code)
}

func TestExistingPut(t *testing.T) {
	if err := run([]string{"put", t.TempDir(), "alpha", "one"}); err != nil {
		t.Fatal(err)
	}
	if err := run(nil); err == nil {
		t.Fatal("missing usage rejection")
	}
}

type processResult struct {
	code           int
	stdout, stderr string
}

func invoke(t *testing.T, directory, input string, args ...string) processResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = directory
	cmd.Env = commandEnvironment()
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("command deadline exceeded: %v; stderr=%s", ctx.Err(), &stderr)
	}
	code := 0
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("command failed to run: %v", err)
		}
		code = exit.ExitCode()
	}
	return processResult{code, stdout.String(), stderr.String()}
}
func assertResult(t *testing.T, result processResult, code int) {
	t.Helper()
	if result.code != code || result.stdout != "" {
		t.Errorf("process status=%d stdout=%q stderr=%q; want status=%d and empty stdout", result.code, result.stdout, result.stderr, code)
	}
	if code == 0 && result.stderr != "" {
		t.Errorf("success stderr=%q; want empty", result.stderr)
	}
	if code != 0 && result.stderr == "" {
		t.Error("failed process omitted diagnostic")
	}
}
func assertBytes(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file %s = %q, %v; want %q", path, got, err, want)
	}
}

func TestPutProcessPreservesGrammarAndStreams(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// A positional directory and text beginning with '-' were valid before flags.
	assertResult(t, invoke(t, root, "", "put", "-directory", "alpha", "-literal"), 0)
	path := filepath.Join(root, "-directory", "alpha")
	assertBytes(t, path, "-literal")
	assertResult(t, invoke(t, root, "", "put", "-directory", "alpha", ""), 0)
	assertBytes(t, path, "")
	assertResult(t, invoke(t, root, "", "put", "-directory", "../escape", "bad"), 2)
	assertBytes(t, path, "")
	if _, err := os.Stat(filepath.Join(root, "escape")); !os.IsNotExist(err) {
		t.Errorf("invalid put escaped root: %v", err)
	}
}

func TestApplyProcess(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	directory := filepath.Join(root, "index")
	assertResult(t, invoke(t, root, "alpha,first\nbeta,\"  comma, text  \"\nalpha,last\n", "apply", "--dir", directory), 0)
	assertBytes(t, filepath.Join(directory, "alpha"), "last")
	assertBytes(t, filepath.Join(directory, "beta"), "  comma, text  ")
	assertResult(t, invoke(t, root, "gamma,accepted\nalpha,accepted replacement\nalpha, \t\nlater,forbidden\n", "apply", "--dir", directory), 2)
	assertBytes(t, filepath.Join(directory, "gamma"), "accepted")
	assertBytes(t, filepath.Join(directory, "alpha"), "accepted replacement")
	if _, err := os.Stat(filepath.Join(directory, "later")); !os.IsNotExist(err) {
		t.Errorf("later row exists: %v", err)
	}
	assertResult(t, invoke(t, root, "delta,accepted\nmalformed\nlater,forbidden\n", "apply", "--dir", directory), 2)
	assertBytes(t, filepath.Join(directory, "delta"), "accepted")
	if _, err := os.Stat(filepath.Join(directory, "later")); !os.IsNotExist(err) {
		t.Errorf("later row exists: %v", err)
	}
}

func TestStartupArgumentsProcess(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"no command", nil},
		{"unknown command", []string{"unknown"}},
		{"put missing text", []string{"put", "index", "alpha"}},
		{"put extra argument", []string{"put", "index", "alpha", "one", "extra"}},
		{"apply missing directory", []string{"apply"}},
		{"apply unknown flag", []string{"apply", "--wrong", "x"}},
		{"apply extra argument", []string{"apply", "--dir", "index", "extra"}},
		{"watch invalid scheme", []string{"watch", "--url", "file://host/x", "--file", "snapshot", "--interval", "1s"}},
		{"watch missing host", []string{"watch", "--url", "http:///records", "--file", "snapshot", "--interval", "1s"}},
		{"watch malformed URL", []string{"watch", "--url", "http://%", "--file", "snapshot", "--interval", "1s"}},
		{"watch missing path", []string{"watch", "--url", "http://127.0.0.1", "--interval", "1s"}},
		{"watch zero interval", []string{"watch", "--url", "http://127.0.0.1", "--file", "snapshot", "--interval", "0s"}},
		{"watch negative interval", []string{"watch", "--url", "http://127.0.0.1", "--file", "snapshot", "--interval", "-1s"}},
		{"watch invalid duration", []string{"watch", "--url", "http://127.0.0.1", "--file", "snapshot", "--interval", "soon"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { t.Parallel(); assertResult(t, invoke(t, t.TempDir(), "", tc.args...), 2) })
	}
}

func TestWatchFirstFailureProcess(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"status", 206, `[{"key":"alpha","text":"valid"}]`},
		{"validation", 200, `[{"key":"alpha","text":" "}]`},
		{"decode", 200, `[][]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(tc.status)
				io.WriteString(w, tc.body)
			}))
			t.Cleanup(server.Close)
			root := t.TempDir()
			path := filepath.Join(root, "snapshot")
			if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
				t.Fatal(err)
			}
			assertResult(t, invoke(t, root, "", "watch", "--url", server.URL, "--file", path, "--interval", "10ms"), 2)
			assertBytes(t, path, "prior")
			if calls.Load() != 1 {
				t.Errorf("first failure made %d requests; want 1", calls.Load())
			}
		})
	}
}

func TestWatchRecurrenceAndSignalsProcess(t *testing.T) {
	for _, signal := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		t.Run(signal.String(), func(t *testing.T) {
			t.Parallel()
			events := make(chan int32, 10)
			secondGate, canceled := make(chan struct{}), make(chan struct{})
			var gateOnce, canceledOnce sync.Once
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				call := calls.Add(1)
				select {
				case events <- call:
				case <-r.Context().Done():
					return
				}
				switch call {
				case 1:
					io.WriteString(w, `[{"key":"alpha","text":"one"}]`)
				case 2:
					select {
					case <-secondGate:
					case <-r.Context().Done():
						return
					}
					io.WriteString(w, `[{"key":"alpha","text":"two"}]`)
				default:
					<-r.Context().Done()
					canceledOnce.Do(func() { close(canceled) })
				}
			}))
			t.Cleanup(server.Close)
			root := t.TempDir()
			path := filepath.Join(root, "snapshot")
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			cmd := exec.CommandContext(ctx, executable, "watch", "--url", server.URL, "--file", path, "--interval", "20ms")
			cmd.Dir, cmd.Env = root, commandEnvironment()
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Start(); err != nil {
				cancel()
				t.Fatal(err)
			}
			done, exited := make(chan error, 1), make(chan struct{})
			go func() { done <- cmd.Wait(); close(exited) }()
			t.Cleanup(func() {
				gateOnce.Do(func() { close(secondGate) })
				cancel()
				select {
				case <-exited:
				case <-time.After(5 * time.Second):
					t.Error("watch process did not finish during cleanup")
				}
			})
			waitRequest := func(want int32) {
				t.Helper()
				select {
				case got := <-events:
					if got != want {
						t.Fatalf("request=%d; want %d", got, want)
					}
				case <-exited:
					t.Fatalf("watch exited early: %v; stderr=%s", <-done, &stderr)
				case <-time.After(5 * time.Second):
					t.Fatalf("watch did not issue request %d", want)
				}
			}
			waitRequest(1)
			waitRequest(2)
			assertBytes(t, path, "[{\"key\":\"alpha\",\"text\":\"one\"}]\n")
			gateOnce.Do(func() { close(secondGate) })
			waitRequest(3)
			assertBytes(t, path, "[{\"key\":\"alpha\",\"text\":\"two\"}]\n")
			if err := cmd.Process.Signal(signal); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				if err != nil {
					t.Fatalf("watch signal exit = %v; stderr=%s", err, &stderr)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("watch did not join after signal")
			}
			select {
			case <-canceled:
			case <-time.After(5 * time.Second):
				t.Fatal("in-flight HTTP request was not canceled")
			}
			assertBytes(t, path, "[{\"key\":\"alpha\",\"text\":\"two\"}]\n")
			assertResult(t, processResult{0, stdout.String(), stderr.String()}, 0)
			if calls.Load() != 3 {
				t.Errorf("requests=%d; want 3", calls.Load())
			}
		})
	}
}

func TestApplyProcessPublicationFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	directory := filepath.Join(root, "index")
	if err := os.MkdirAll(filepath.Join(directory, "blocked"), 0700); err != nil {
		t.Fatal(err)
	}
	assertResult(t, invoke(t, root, "alpha,accepted\nblocked,rejected\nlater,forbidden\n", "apply", "--dir", directory), 2)
	assertBytes(t, filepath.Join(directory, "alpha"), "accepted")
	if _, err := os.Stat(filepath.Join(directory, "later")); !os.IsNotExist(err) {
		t.Errorf("later row exists: %v", err)
	}
}

func TestWatchLaterFailureProcess(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			io.WriteString(w, `[{"key":"alpha","text":"accepted"}]`)
			return
		}
		w.WriteHeader(http.StatusPartialContent)
		io.WriteString(w, `[{"key":"alpha","text":"rejected"}]`)
	}))
	t.Cleanup(server.Close)
	root := t.TempDir()
	path := filepath.Join(root, "snapshot")
	if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
		t.Fatal(err)
	}
	assertResult(t, invoke(t, root, "", "watch", "--url", server.URL, "--file", path, "--interval", "10ms"), 2)
	assertBytes(t, path, "[{\"key\":\"alpha\",\"text\":\"accepted\"}]\n")
	if calls.Load() != 2 {
		t.Errorf("requests=%d; want two", calls.Load())
	}
}

func TestWatchInvalidStartupMakesNoRequest(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); io.WriteString(w, "[]") }))
	t.Cleanup(server.Close)
	root := t.TempDir()
	path := filepath.Join(root, "snapshot")
	if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, interval := range []string{"0", "-1s", "soon"} {
		assertResult(t, invoke(t, root, "", "watch", "--url", server.URL, "--file", path, "--interval", interval), 2)
		assertBytes(t, path, "prior")
	}
	if calls.Load() != 0 {
		t.Errorf("invalid startup issued %d HTTP requests", calls.Load())
	}
}
