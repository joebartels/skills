//go:build darwin || linux || freebsd

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestWatchFirstRefreshFailureProcess(t *testing.T) {
	for _, tc := range []struct {
		name, data string
		status     int
	}{
		{"status", `[]`, 503},
		{"invalid record", `[{"key":"A","text":"bad"}]`, 200},
		{"trailing data", `[] []`, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := localServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("method = %s", r.Method)
				}
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.data)
			}))
			path := filepath.Join(t.TempDir(), "records.json")
			if err := os.WriteFile(path, []byte("prior\n"), 0600); err != nil {
				t.Fatal(err)
			}
			result := invoke(t, t.TempDir(), "", "watch", "--url", server.URL, "--file", path, "--interval", "30ms")
			assertProcess(t, result, 2)
			assertBytes(t, path, "prior\n")
		})
	}
}

func TestWatchRecursAndJoinsOnSignals(t *testing.T) {
	for _, signal := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		t.Run(signal.String(), func(t *testing.T) {
			requests := make(chan int, 16)
			var count atomic.Int32
			server := localServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("method = %s", r.Method)
				}
				n := int(count.Add(1))
				fmt.Fprintf(w, `[{"key":"alpha","text":"cycle %d"}]`, n)
				select {
				case requests <- n:
				default:
				}
			}))
			path := filepath.Join(t.TempDir(), "records.json")
			process := startWatch(t, server.URL, path, "150ms")
			if got := awaitEvent(t, requests); got != 1 {
				t.Fatalf("first request = %d", got)
			}
			waitForFile(t, path, "[{\"key\":\"alpha\",\"text\":\"cycle 1\"}]\n", process)
			if got := awaitEvent(t, requests); got != 2 {
				t.Fatalf("second request = %d", got)
			}
			waitForFile(t, path, "[{\"key\":\"alpha\",\"text\":\"cycle 2\"}]\n", process)
			if err := process.command.Process.Signal(signal); err != nil {
				t.Fatal(err)
			}
			assertProcess(t, process.result(t), 0)
			before := count.Load()
			select {
			case <-time.After(200 * time.Millisecond):
			}
			if after := count.Load(); after != before {
				t.Fatalf("requests continued after process exit: %d -> %d", before, after)
			}
		})
	}
}

func TestWatchCancelsInFlightResponse(t *testing.T) {
	started, canceled := make(chan struct{}), make(chan struct{})
	server := localServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"key":"alpha","text":"`)
		w.(http.Flusher).Flush()
		close(started)
		select {
		case <-r.Context().Done():
			close(canceled)
		case <-time.After(5 * time.Second):
			t.Error("request cancellation deadline exceeded")
		}
	}))
	path := filepath.Join(t.TempDir(), "records.json")
	if err := os.WriteFile(path, []byte("prior exact bytes\n"), 0600); err != nil {
		t.Fatal(err)
	}
	process := startWatch(t, server.URL, path, "30ms")
	awaitEvent(t, started)
	if err := process.command.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	assertProcess(t, process.result(t), 0)
	awaitEvent(t, canceled)
	assertBytes(t, path, "prior exact bytes\n")
}

func localServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("real HTTP listener boundary: %v", err)
	}
	server := httptest.NewUnstartedServer(handler)
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)
	return server
}

type runningProcess struct {
	command        *exec.Cmd
	stdout, stderr bytes.Buffer
	done           chan struct{}
	err            error
}

func startWatch(t *testing.T, endpoint, path, interval string) *runningProcess {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	process := &runningProcess{command: exec.CommandContext(ctx, executable, "watch", "--url", endpoint, "--file", path, "--interval", interval), done: make(chan struct{})}
	process.command.Stdout, process.command.Stderr = &process.stdout, &process.stderr
	if err := process.command.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	go func() { process.err = process.command.Wait(); close(process.done) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-process.done:
		default:
			process.command.Process.Kill()
			select {
			case <-process.done:
			case <-time.After(time.Second):
				t.Error("child did not join during test cleanup")
			}
		}
	})
	return process
}
func (p *runningProcess) result(t *testing.T) processResult {
	t.Helper()
	awaitEvent(t, p.done)
	status := 0
	if p.err != nil {
		var exitErr *exec.ExitError
		if !errors.As(p.err, &exitErr) {
			t.Fatalf("watch process: %v", p.err)
		}
		status = exitErr.ExitCode()
	}
	return processResult{status, p.stdout.String(), p.stderr.String()}
}
func waitForFile(t *testing.T, path, want string, p *runningProcess) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		got, err := os.ReadFile(path)
		if err == nil && string(got) == want {
			return
		}
		select {
		case <-p.done:
			t.Fatalf("watch exited before expected snapshot: %+v", p.result(t))
		case <-deadline.C:
			t.Fatalf("snapshot deadline: got %q, %v; want %q", got, err, want)
		case <-ticker.C:
		}
	}
}
func awaitEvent[T any](t *testing.T, event <-chan T) T {
	t.Helper()
	select {
	case value := <-event:
		return value
	case <-time.After(3 * time.Second):
		t.Fatal("process/lifecycle event deadline exceeded")
	}
	var zero T
	return zero
}
