package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestCLIProcessContracts(t *testing.T) {
	binary := buildCommand(t)
	t.Run("put remains positional", func(t *testing.T) {
		root := t.TempDir()
		for _, text := range []string{"one", "", "  "} {
			code, stdout, stderr := commandResult(t, binary, root, "", "put", "-records", "alpha", text)
			assertProcess(t, code, stdout, stderr, 0)
			assertCommandFile(t, filepath.Join(root, "-records", "alpha"), text)
		}
		code, stdout, stderr := commandResult(t, binary, root, "", "put", "-records", "../escape", "bad")
		assertProcess(t, code, stdout, stderr, 2)
		assertCommandFile(t, filepath.Join(root, "-records", "alpha"), "  ")
	})
	t.Run("put usage", func(t *testing.T) {
		for _, args := range [][]string{nil, {"put"}, {"put", t.TempDir(), "alpha", "text", "extra"}, {"unknown"}} {
			code, stdout, stderr := commandResult(t, binary, t.TempDir(), "", args...)
			assertProcess(t, code, stdout, stderr, 2)
		}
	})
	t.Run("apply success", func(t *testing.T) {
		root := t.TempDir()
		input := "alpha,one\nbeta,\"a,b\nnext\"\nalpha, two \n"
		code, stdout, stderr := commandResult(t, binary, t.TempDir(), input, "apply", "--dir", root)
		assertProcess(t, code, stdout, stderr, 0)
		assertCommandFile(t, filepath.Join(root, "alpha"), " two ")
		assertCommandFile(t, filepath.Join(root, "beta"), "a,b\nnext")
	})
	t.Run("apply prefix after validation", func(t *testing.T) {
		root := t.TempDir()
		code, stdout, stderr := commandResult(t, binary, t.TempDir(), "alpha,accepted\nalpha, \nlater,forbidden\n", "apply", "--dir", root)
		assertProcess(t, code, stdout, stderr, 2)
		assertCommandFile(t, filepath.Join(root, "alpha"), "accepted")
		assertCommandAbsent(t, filepath.Join(root, "later"))
	})
	t.Run("apply prefix after parse", func(t *testing.T) {
		root := t.TempDir()
		code, stdout, stderr := commandResult(t, binary, t.TempDir(), "alpha,accepted\nbeta,too,many\nlater,forbidden\n", "apply", "--dir", root)
		assertProcess(t, code, stdout, stderr, 2)
		assertCommandFile(t, filepath.Join(root, "alpha"), "accepted")
		assertCommandAbsent(t, filepath.Join(root, "beta"))
		assertCommandAbsent(t, filepath.Join(root, "later"))
	})
	t.Run("apply prefix after publication", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "beta"), 0700); err != nil {
			t.Fatal(err)
		}
		code, stdout, stderr := commandResult(t, binary, t.TempDir(), "alpha,accepted\nbeta,rejected\nlater,forbidden\n", "apply", "--dir", root)
		assertProcess(t, code, stdout, stderr, 2)
		assertCommandFile(t, filepath.Join(root, "alpha"), "accepted")
		assertCommandAbsent(t, filepath.Join(root, "later"))
	})
	t.Run("startup rejects", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "snapshot")
		for _, args := range [][]string{
			{"apply"}, {"apply", "--dir", t.TempDir(), "extra"}, {"apply", "--unknown"},
			{"watch"},
			{"watch", "--url", "ftp://example.test", "--file", path, "--interval", "1s"},
			{"watch", "--url", "http:///missing", "--file", path, "--interval", "1s"},
			{"watch", "--url", "http://example.test", "--file", "", "--interval", "1s"},
			{"watch", "--url", "http://example.test", "--file", path, "--interval", "0s"},
			{"watch", "--url", "http://example.test", "--file", path, "--interval", "-1s"},
			{"watch", "--url", "http://example.test", "--file", path, "--interval", "bad"},
			{"watch", "--url", "http://example.test", "--file", path, "--interval", "1s", "extra"},
		} {
			code, stdout, stderr := commandResult(t, binary, t.TempDir(), "", args...)
			assertProcess(t, code, stdout, stderr, 2)
		}
	})
	t.Run("failed first refresh", func(t *testing.T) {
		for _, tc := range []struct {
			name, body string
			status     int
		}{
			{"neighbor status", `[{"key":"alpha","text":"new"}]`, 206},
			{"decode", `[{`, 200},
			{"validation", `[{"key":"alpha","text":""}]`, 200},
		} {
			t.Run(tc.name, func(t *testing.T) {
				var requests atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests.Add(1)
					w.WriteHeader(tc.status)
					io.WriteString(w, tc.body)
				}))
				t.Cleanup(server.Close)
				path := filepath.Join(t.TempDir(), "snapshot")
				if err := os.WriteFile(path, []byte("old bytes"), 0600); err != nil {
					t.Fatal(err)
				}
				code, stdout, stderr := commandResult(t, binary, t.TempDir(), "", "watch", "--url", server.URL, "--file", path, "--interval", "20ms")
				assertProcess(t, code, stdout, stderr, 2)
				assertCommandFile(t, path, "old bytes")
				if requests.Load() != 1 {
					t.Fatalf("requests = %d; want only first failure", requests.Load())
				}
			})
		}
	})
	t.Run("failed first publication", func(t *testing.T) {
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			io.WriteString(w, `[{"key":"alpha","text":"new"}]`)
		}))
		t.Cleanup(server.Close)
		path := filepath.Join(t.TempDir(), "snapshot")
		if err := os.Mkdir(path, 0700); err != nil {
			t.Fatal(err)
		}
		marker := filepath.Join(path, "marker")
		if err := os.WriteFile(marker, []byte("old"), 0600); err != nil {
			t.Fatal(err)
		}
		code, stdout, stderr := commandResult(t, binary, t.TempDir(), "", "watch", "--url", server.URL, "--file", path, "--interval", "20ms")
		assertProcess(t, code, stdout, stderr, 2)
		assertCommandFile(t, marker, "old")
		if requests.Load() != 1 {
			t.Fatalf("requests = %d; want only first publication failure", requests.Load())
		}
	})
	t.Run("watch recurrence and termination", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("process interrupt test exercises POSIX signal handling")
		}
		var requests atomic.Int32
		secondStarted, allowSecond := make(chan struct{}), make(chan struct{})
		var unblock sync.Once
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n := requests.Add(1)
			if n == 2 {
				close(secondStarted)
				select {
				case <-allowSecond:
				case <-r.Context().Done():
					return
				}
			}
			if n > 2 {
				<-r.Context().Done()
				return
			}
			fmt.Fprintf(w, `[{"key":"alpha","text":"cycle %d"}]`, n)
		}))
		t.Cleanup(server.Close)
		t.Cleanup(func() { unblock.Do(func() { close(allowSecond) }) })
		path := filepath.Join(t.TempDir(), "snapshot")
		cmd, stdout, stderr, finished, result := startWatch(t, binary, server.URL, path, "50ms")
		waitCommandFile(t, path, `[{"key":"alpha","text":"cycle 1"}]`+"\n")
		waitCommandClosed(t, secondStarted, "second successful refresh start")
		unblock.Do(func() { close(allowSecond) })
		waitCommandFile(t, path, `[{"key":"alpha","text":"cycle 2"}]`+"\n")
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
			t.Fatal(err)
		}
		waitCommandClosed(t, finished, "watch termination")
		if err := <-result; err != nil {
			t.Fatalf("watch wait: %v\nstdout=%q stderr=%q", err, stdout, stderr)
		}
		assertProcess(t, 0, stdout.String(), stderr.String(), 0)
		if requests.Load() < 2 {
			t.Fatalf("requests = %d; want successful recurrence", requests.Load())
		}
	})
	t.Run("watch joins canceled network request", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("process interrupt test exercises POSIX signal handling")
		}
		started, handlerDone := make(chan struct{}), make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(started)
			<-r.Context().Done()
			close(handlerDone)
		}))
		t.Cleanup(server.Close)
		path := filepath.Join(t.TempDir(), "snapshot")
		if err := os.WriteFile(path, []byte("old bytes"), 0600); err != nil {
			t.Fatal(err)
		}
		cmd, stdout, stderr, finished, result := startWatch(t, binary, server.URL, path, "20ms")
		waitCommandClosed(t, started, "watch HTTP request")
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			t.Fatal(err)
		}
		waitCommandClosed(t, finished, "watch cancellation")
		if err := <-result; err != nil {
			t.Fatalf("watch wait: %v\nstdout=%q stderr=%q", err, stdout, stderr)
		}
		waitCommandClosed(t, handlerDone, "server observed request cancellation")
		assertProcess(t, 0, stdout.String(), stderr.String(), 0)
		assertCommandFile(t, path, "old bytes")
	})
}

func buildCommand(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "indexer")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	// Build inherits the caller's normal Go cache and toolchain settings; the
	// executable below receives a separate explicit application environment.
	cmd.Env = os.Environ()
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build indexer: %v\n%s", err, output)
	}
	return binary
}

func commandEnv() []string {
	var env []string
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "SYSTEMROOT"} {
		if value, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+value)
		}
	}
	return env
}

func commandResult(t *testing.T, binary, dir, stdin string, args ...string) (int, string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir, cmd.Env, cmd.Stdin = dir, commandEnv(), strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("command exceeded deadline: %v", ctx.Err())
	}
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatalf("start command: %v", err)
		}
		code = exit.ExitCode()
	}
	return code, stdout.String(), stderr.String()
}

func startWatch(t *testing.T, binary, endpoint, path, interval string) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer, <-chan struct{}, <-chan error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	cmd := exec.CommandContext(ctx, binary, "watch", "--url", endpoint, "--file", path, "--interval", interval)
	cmd.Env = commandEnv()
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	finished, result := make(chan struct{}), make(chan error, 1)
	go func() { defer close(finished); result <- cmd.Wait() }()
	t.Cleanup(func() {
		cancel()
		waitCommandClosed(t, finished, "watch process cleanup")
	})
	return cmd, stdout, stderr, finished, result
}

func assertProcess(t *testing.T, code int, stdout, stderr string, want int) {
	t.Helper()
	if code != want || stdout != "" || (want == 0 && stderr != "") || (want != 0 && stderr == "") {
		t.Fatalf("process = status %d, stdout %q, stderr %q; want status %d and conventional streams", code, stdout, stderr, want)
	}
}

func assertCommandFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file = %q, %v; want %q", got, err, want)
	}
}

func assertCommandAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("later file %q: %v; want absent", path, err)
	}
}

func waitCommandFile(t *testing.T, path, want string) {
	t.Helper()
	deadline := time.NewTimer(3 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		got, err := os.ReadFile(path)
		if err == nil && string(got) == want {
			return
		}
		select {
		case <-deadline.C:
			t.Fatalf("did not observe snapshot %q; last bytes %q, error %v", want, got, err)
		case <-ticker.C:
		}
	}
}

func waitCommandClosed(t *testing.T, done <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}
