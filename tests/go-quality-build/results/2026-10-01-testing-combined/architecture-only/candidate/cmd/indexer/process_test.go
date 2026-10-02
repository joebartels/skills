package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func buildCommand(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "indexer")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", path, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build command: %v\n%s", err, output)
	}
	return path
}

func checkCommand(t *testing.T, binary string, args []string, input string, wantCode int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("command %q exceeded deadline: %v", args, ctx.Err())
	}
	gotCode := 0
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("command %q could not run: %v", args, err)
		}
		gotCode = exit.ExitCode()
	}
	if gotCode != wantCode || stdout.Len() != 0 || (wantCode == 0 && stderr.Len() != 0) || (wantCode != 0 && stderr.Len() == 0) {
		t.Fatalf("command %q: exit=%d stdout=%q stderr=%q; want exit=%d, no stdout, appropriate stderr", args, gotCode, stdout.String(), stderr.String(), wantCode)
	}
}

func checkFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("ReadFile(%q) = %q, %v; want %q", path, got, err, want)
	}
}

func TestCommandPutApplyAndUsage(t *testing.T) {
	binary := buildCommand(t)
	t.Run("put existing positional grammar", func(t *testing.T) {
		root := t.TempDir()
		checkCommand(t, binary, []string{"put", root, "alpha", "-literal"}, "", 0)
		checkFile(t, filepath.Join(root, "alpha"), "-literal")
		checkCommand(t, binary, []string{"put", root, "alpha", ""}, "", 0)
		checkFile(t, filepath.Join(root, "alpha"), "")
		checkCommand(t, binary, []string{"put", root, "Alpha", "bad"}, "", 2)
		checkFile(t, filepath.Join(root, "alpha"), "")
	})
	t.Run("apply exact text and duplicate replacement", func(t *testing.T) {
		root := t.TempDir()
		checkCommand(t, binary, []string{"apply", "--dir", root}, "alpha,one\nbeta,\" text,with\nnewline \"\nalpha,two\n", 0)
		checkFile(t, filepath.Join(root, "alpha"), "two")
		checkFile(t, filepath.Join(root, "beta"), " text,with\nnewline ")
	})
	for _, rejected := range []string{"alpha,\" \t\"\n", "alpha\n", "alpha,three,fields\n"} {
		t.Run("apply rejected prefix "+fmt.Sprintf("%q", rejected), func(t *testing.T) {
			root := t.TempDir()
			checkCommand(t, binary, []string{"apply", "--dir", root}, "alpha,accepted\n"+rejected+"gamma,later\n", 2)
			checkFile(t, filepath.Join(root, "alpha"), "accepted")
			if _, err := os.Stat(filepath.Join(root, "gamma")); !os.IsNotExist(err) {
				t.Fatalf("later row was written: %v", err)
			}
		})
	}
	t.Run("apply write failure", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "beta"), 0700); err != nil {
			t.Fatal(err)
		}
		checkCommand(t, binary, []string{"apply", "--dir", root}, "alpha,accepted\nbeta,rejected\ngamma,later\n", 2)
		checkFile(t, filepath.Join(root, "alpha"), "accepted")
		if _, err := os.Stat(filepath.Join(root, "gamma")); !os.IsNotExist(err) {
			t.Fatalf("later row was written: %v", err)
		}
	})
	invalid := [][]string{
		nil, {"unknown"}, {"put"}, {"put", "a", "b", "c", "extra"},
		{"apply"}, {"apply", "--bogus"}, {"apply", "--dir", t.TempDir(), "extra"},
		{"watch"}, {"watch", "--url", "ftp://host/x", "--file", "x", "--interval", "1s"},
		{"watch", "--url", "http:///missing-host", "--file", "x", "--interval", "1s"},
		{"watch", "--url", "http://:80", "--file", "x", "--interval", "1s"},
		{"watch", "--url", "https://example.invalid", "--file", "", "--interval", "1s"},
		{"watch", "--url", "https://example.invalid", "--file", "x", "--interval", "0"},
		{"watch", "--url", "https://example.invalid", "--file", "x", "--interval", "-1s"},
		{"watch", "--url", "https://example.invalid", "--file", "x", "--interval", "bad"},
		{"watch", "--url", "https://example.invalid", "--file", "x"},
		{"watch", "--url", "https://example.invalid", "--file", "x", "--interval", "1s", "extra"},
	}
	for n, args := range invalid {
		t.Run(fmt.Sprintf("invalid startup %d", n), func(t *testing.T) { checkCommand(t, binary, args, "", 2) })
	}
}

func localServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("local HTTP listener required for command integration: %v", err)
	}
	server := &httptest.Server{Listener: listener, Config: &http.Server{Handler: handler}}
	server.Start()
	t.Cleanup(server.Close)
	return server
}

type runningCommand struct {
	cmd            *exec.Cmd
	done           chan struct{}
	err            error
	stdout, stderr bytes.Buffer
}

func startWatch(t *testing.T, binary, endpoint, path, interval string) *runningCommand {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	run := &runningCommand{cmd: exec.CommandContext(ctx, binary, "watch", "--url", endpoint, "--file", path, "--interval", interval), done: make(chan struct{})}
	run.cmd.Stdout, run.cmd.Stderr = &run.stdout, &run.stderr
	if err := run.cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	go func() { run.err = run.cmd.Wait(); close(run.done) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-run.done:
		case <-time.After(2 * time.Second):
			t.Error("watch process did not join after cleanup cancellation")
		}
	})
	return run
}

func waitProcess(t *testing.T, run *runningCommand, wantCode int) {
	t.Helper()
	select {
	case <-run.done:
	case <-time.After(3 * time.Second):
		t.Fatal("watch process exceeded completion deadline")
	}
	code := 0
	if run.err != nil {
		exit, ok := run.err.(*exec.ExitError)
		if !ok {
			t.Fatalf("watch wait: %v", run.err)
		}
		code = exit.ExitCode()
	}
	if code != wantCode || run.stdout.Len() != 0 || (wantCode == 0 && run.stderr.Len() != 0) || (wantCode != 0 && run.stderr.Len() == 0) {
		t.Fatalf("watch: exit=%d stdout=%q stderr=%q; want %d", code, run.stdout.String(), run.stderr.String(), wantCode)
	}
}

func waitPublished(t *testing.T, path, want string, run *runningCommand) {
	t.Helper()
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		if got, err := os.ReadFile(path); err == nil && string(got) == want {
			return
		}
		select {
		case <-run.done:
			t.Fatalf("watch exited before publication: %v stderr=%q", run.err, run.stderr.String())
		case <-timeout.C:
			t.Fatalf("watch did not publish %q", want)
		case <-tick.C:
		}
	}
}

func TestCommandWatchLifecycle(t *testing.T) {
	binary := buildCommand(t)
	for _, sig := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		t.Run("recurs and joins on "+sig.String(), func(t *testing.T) {
			var calls atomic.Int32
			server := localServer(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("request method = %s", r.Method)
				}
				n := calls.Add(1)
				fmt.Fprintf(w, `[{"key":"alpha","text":"%d"}]`, n)
			})
			path := filepath.Join(t.TempDir(), "records.json")
			run := startWatch(t, binary, server.URL, path, "150ms")
			waitPublished(t, path, "[{\"key\":\"alpha\",\"text\":\"1\"}]\n", run)
			waitPublished(t, path, "[{\"key\":\"alpha\",\"text\":\"2\"}]\n", run)
			if err := run.cmd.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			waitProcess(t, run, 0)
			if calls.Load() < 2 {
				t.Fatalf("HTTP calls = %d; recurrence not exercised", calls.Load())
			}
			data, err := os.ReadFile(path)
			var records []struct{ Key, Text string }
			if err != nil || json.Unmarshal(data, &records) != nil || len(records) != 1 || records[0].Key != "alpha" {
				t.Fatalf("published snapshot = %q, %v", data, err)
			}
			n, err := strconv.Atoi(records[0].Text)
			if err != nil || n < 2 || int32(n) > calls.Load() {
				t.Fatalf("published record = %q; HTTP calls = %d", records[0].Text, calls.Load())
			}
		})
	}
	t.Run("interrupt cancels in flight request", func(t *testing.T) {
		started, canceled := make(chan struct{}), make(chan struct{})
		server := localServer(t, func(w http.ResponseWriter, r *http.Request) {
			close(started)
			select {
			case <-r.Context().Done():
				close(canceled)
			case <-time.After(4 * time.Second):
				t.Error("request context was not canceled")
			}
		})
		path := filepath.Join(t.TempDir(), "records.json")
		if err := os.WriteFile(path, []byte("prior bytes"), 0600); err != nil {
			t.Fatal(err)
		}
		run := startWatch(t, binary, server.URL, path, "20ms")
		select {
		case <-started:
		case <-run.done:
			t.Fatalf("watch exited before fetch: %v", run.err)
		case <-time.After(2 * time.Second):
			t.Fatal("watch did not start request")
		}
		if err := run.cmd.Process.Signal(os.Interrupt); err != nil {
			t.Fatal(err)
		}
		waitProcess(t, run, 0)
		select {
		case <-canceled:
		case <-time.After(2 * time.Second):
			t.Fatal("watch did not cancel HTTP request context")
		}
		checkFile(t, path, "prior bytes")
	})
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"first status failure", "[]", 503},
		{"first decode failure", "[] []", 200},
		{"first validation failure", `[{"key":"Alpha","text":"one"}]`, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			server := localServer(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			})
			path := filepath.Join(t.TempDir(), "records.json")
			if err := os.WriteFile(path, []byte("prior bytes"), 0600); err != nil {
				t.Fatal(err)
			}
			checkCommand(t, binary, []string{"watch", "--url", server.URL, "--file", path, "--interval", "20ms"}, "", 2)
			checkFile(t, path, "prior bytes")
			if calls.Load() != 1 {
				t.Fatalf("HTTP calls = %d; want one failed first refresh", calls.Load())
			}
		})
	}
	t.Run("later failure retains last successful refresh", func(t *testing.T) {
		var calls atomic.Int32
		server := localServer(t, func(w http.ResponseWriter, r *http.Request) {
			if calls.Add(1) == 1 {
				fmt.Fprint(w, `[{"key":"alpha","text":"accepted"}]`)
				return
			}
			w.WriteHeader(http.StatusServiceUnavailable)
		})
		path := filepath.Join(t.TempDir(), "records.json")
		checkCommand(t, binary, []string{"watch", "--url", server.URL, "--file", path, "--interval", "20ms"}, "", 2)
		checkFile(t, path, "[{\"key\":\"alpha\",\"text\":\"accepted\"}]\n")
		if calls.Load() != 2 {
			t.Fatalf("HTTP calls = %d; want success then failure", calls.Load())
		}
	})
}
