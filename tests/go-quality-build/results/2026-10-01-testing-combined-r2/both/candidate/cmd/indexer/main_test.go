package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

var executable string

func TestMain(m *testing.M) {
	root, err := os.MkdirTemp("", "indexer-command-tests-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	executable = filepath.Join(root, "indexer")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	command := exec.CommandContext(ctx, "go", "build", "-o", executable, ".")
	command.Env = processEnvironment()
	output, err := command.CombinedOutput()
	cancel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build command: %v\n%s", err, output)
		os.RemoveAll(root)
		os.Exit(1)
	}
	status := m.Run()
	os.RemoveAll(root)
	os.Exit(status)
}

func TestExistingPut(t *testing.T) {
	if err := run([]string{"put", t.TempDir(), "alpha", "one"}); err != nil {
		t.Fatal(err)
	}
	if err := run(nil); err == nil {
		t.Fatal("missing usage rejection")
	}
}

func TestPutProcessStreamsAndCompatibility(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, text := range []string{"first", "-dash positional text", ""} {
		out, stderr, code := runProcess(t, []string{"put", root, "alpha", text}, "")
		assertOutcome(t, out, stderr, code, 0)
		assertPath(t, filepath.Join(root, "alpha"), text)
	}
	out, stderr, code := runProcess(t, []string{"put", root, "../escape", "bad"}, "")
	assertOutcome(t, out, stderr, code, 2)
	assertPath(t, filepath.Join(root, "alpha"), "")
}

func TestApplyProcess(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	out, stderr, code := runProcess(t, []string{"apply", "--dir", root}, "alpha,first\nbeta,\"comma, text\"\nalpha,replacement\n")
	assertOutcome(t, out, stderr, code, 0)
	assertPath(t, filepath.Join(root, "alpha"), "replacement")
	assertPath(t, filepath.Join(root, "beta"), "comma, text")
	out, stderr, code = runProcess(t, []string{"apply", "--dir", root}, "alpha,accepted\nalpha,  \ngamma,later\n")
	assertOutcome(t, out, stderr, code, 2)
	assertPath(t, filepath.Join(root, "alpha"), "accepted")
	assertPath(t, filepath.Join(root, "beta"), "comma, text")
	if data, err := os.ReadFile(filepath.Join(root, "gamma")); !os.IsNotExist(err) {
		t.Fatalf("later record = %q, %v", data, err)
	}
}

func TestApplyProcessWriteFailureRetainsPrefix(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "beta"), 0700); err != nil {
		t.Fatal(err)
	}
	out, stderr, code := runProcess(t, []string{"apply", "--dir", root}, "alpha,prefix\nbeta,rejected\ngamma,later\n")
	assertOutcome(t, out, stderr, code, 2)
	assertPath(t, filepath.Join(root, "alpha"), "prefix")
	if data, err := os.ReadFile(filepath.Join(root, "gamma")); !os.IsNotExist(err) {
		t.Fatalf("later record = %q, %v", data, err)
	}
}

func TestStartupRejectionsAreProcessFailures(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"missing command", nil}, {"unknown command", []string{"unknown"}}, {"put usage", []string{"put", "dir"}},
		{"apply missing directory", []string{"apply"}}, {"apply extra argument", []string{"apply", "--dir", "unused", "extra"}},
		{"apply unknown option", []string{"apply", "--unknown"}},
		{"watch missing values", []string{"watch"}},
		{"watch wrong scheme", []string{"watch", "--url", "ftp://example.test", "--file", "unused", "--interval", "1s"}},
		{"watch missing host", []string{"watch", "--url", "http:///source", "--file", "unused", "--interval", "1s"}},
		{"watch empty hostname", []string{"watch", "--url", "http://:123/source", "--file", "unused", "--interval", "1s"}},
		{"watch missing file", []string{"watch", "--url", "http://example.test", "--interval", "1s"}},
		{"watch malformed interval", []string{"watch", "--url", "http://example.test", "--file", "unused", "--interval", "bad"}},
		{"watch zero interval", []string{"watch", "--url", "http://example.test", "--file", "unused", "--interval", "0s"}},
		{"watch negative interval", []string{"watch", "--url", "http://example.test", "--file", "unused", "--interval", "-1s"}},
		{"watch extra argument", []string{"watch", "--url", "http://example.test", "--file", "unused", "--interval", "1s", "extra"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, stderr, code := runProcess(t, tc.args, "")
			assertOutcome(t, out, stderr, code, 2)
		})
	}
}

func TestWatchFirstRefreshFailure(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1); http.Error(w, "unavailable", 503) }))
	t.Cleanup(server.Close)
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("prior\x00"), 0600); err != nil {
		t.Fatal(err)
	}
	out, stderr, code := runProcess(t, []string{"watch", "--url", server.URL, "--file", path, "--interval", "10ms"}, "")
	assertOutcome(t, out, stderr, code, 2)
	if requests.Load() != 1 {
		t.Fatalf("first-failure requests = %d", requests.Load())
	}
	assertPath(t, path, "prior\x00")
}

func TestWatchRecursAndTerminatesAfterSuccess(t *testing.T) {
	t.Parallel()
	requests := make(chan int, 16)
	var count atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(count.Add(1))
		io.WriteString(w, fmt.Sprintf(`[{"key":"alpha","text":"cycle %d"}]`, n))
		requests <- n
	}))
	t.Cleanup(server.Close)
	path := filepath.Join(t.TempDir(), "snapshot")
	process := startWatch(t, server.URL, path, "400ms")
	awaitRequest(t, requests, 1)
	awaitRequest(t, requests, 2)
	want := "[{\"key\":\"alpha\",\"text\":\"cycle 2\"}]\n"
	waitForPath(t, path, want, process.done)
	if err := process.command.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	err := awaitProcess(t, process.done)
	process.joined = true
	if err != nil {
		t.Fatalf("watch termination = %v\nstderr: %s", err, process.stderr.String())
	}
	assertOutcome(t, process.stdout.String(), process.stderr.String(), 0, 0)
	assertPath(t, path, want)
	// Decode independently as a consumer, alongside the known-byte assertion.
	var records []struct {
		Key  string `json:"key"`
		Text string `json:"text"`
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &records); err != nil || len(records) != 1 || records[0].Key != "alpha" || records[0].Text != "cycle 2" {
		t.Fatalf("consumer records = %+v, %v", records, err)
	}
}

func TestWatchInterruptCancelsInflightHTTPAndRetainsPriorFile(t *testing.T) {
	t.Parallel()
	started, stopped := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-r.Context().Done(); close(stopped) }))
	t.Cleanup(server.Close)
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("prior exact bytes\x00\n"), 0600); err != nil {
		t.Fatal(err)
	}
	process := startWatch(t, server.URL, path, "10ms")
	awaitChannel(t, started, "inflight HTTP startup")
	if err := process.command.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	err := awaitProcess(t, process.done)
	process.joined = true
	if err != nil {
		t.Fatalf("watch interrupt = %v\nstderr: %s", err, process.stderr.String())
	}
	awaitChannel(t, stopped, "server request cancellation")
	assertOutcome(t, process.stdout.String(), process.stderr.String(), 0, 0)
	assertPath(t, path, "prior exact bytes\x00\n")
}

type watchProcess struct {
	command        *exec.Cmd
	stdout, stderr bytes.Buffer
	done           chan error
	cancel         context.CancelFunc
	joined         bool
}

func startWatch(t *testing.T, endpoint, path, interval string) *watchProcess {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	p := &watchProcess{done: make(chan error, 1), cancel: cancel}
	p.command = exec.CommandContext(ctx, executable, "watch", "--url", endpoint, "--file", path, "--interval", interval)
	p.command.Env = processEnvironment()
	p.command.Stdout, p.command.Stderr = &p.stdout, &p.stderr
	if err := p.command.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cancel()
		if !p.joined {
			select {
			case <-p.done:
				p.joined = true
			case <-time.After(3 * time.Second):
				t.Error("watch process did not join during cleanup")
			}
		}
	})
	go func() { p.done <- p.command.Wait() }()
	return p
}
func runProcess(t *testing.T, args []string, input string) (string, string, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, args...)
	command.Env = processEnvironment()
	command.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("command %q exceeded deadline: %v", args, ctx.Err())
	}
	code := 0
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("command %q could not run: %v", args, err)
		}
		code = exit.ExitCode()
	}
	return stdout.String(), stderr.String(), code
}
func processEnvironment() []string {
	result := []string{}
	for _, name := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOTOOLCHAIN"} {
		if value, ok := os.LookupEnv(name); ok {
			result = append(result, name+"="+value)
		}
	}
	return result
}
func assertOutcome(t *testing.T, stdout, stderr string, code, want int) {
	t.Helper()
	if code != want || stdout != "" {
		t.Fatalf("status/stdout/stderr = %d/%q/%q; want status %d and empty stdout", code, stdout, stderr, want)
	}
	if want == 0 && stderr != "" {
		t.Fatalf("success stderr = %q", stderr)
	}
	if want != 0 && stderr == "" {
		t.Fatal("failure has no stderr diagnostic")
	}
}
func assertPath(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil || string(data) != want {
		t.Fatalf("file = %q, %v; want %q", data, err, want)
	}
}
func awaitProcess(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for command")
		return nil
	}
}
func awaitChannel(t *testing.T, ch <-chan struct{}, label string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %s", label)
	}
}
func awaitRequest(t *testing.T, ch <-chan int, want int) {
	t.Helper()
	select {
	case got := <-ch:
		if got != want {
			t.Fatalf("request %d; want %d", got, want)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for request %d", want)
	}
}
func waitForPath(t *testing.T, path, want string, done chan error) {
	t.Helper()
	deadline := time.NewTimer(3 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		data, err := os.ReadFile(path)
		if err == nil && string(data) == want {
			return
		}
		select {
		case err := <-done:
			done <- err
			t.Fatalf("watch returned before publication: %v", err)
		case <-deadline.C:
			t.Fatalf("timed out waiting for publication; last bytes %q, %v", data, err)
		case <-ticker.C:
		}
	}
}
