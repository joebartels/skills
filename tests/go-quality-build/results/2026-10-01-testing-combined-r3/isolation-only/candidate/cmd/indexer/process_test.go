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
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestCommandProcesses(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "indexer")
	buildCtx, cancelBuild := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancelBuild()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binary, ".")
	// Retain the builder's HOME/cache/toolchain configuration. Application
	// processes below get a small explicit environment instead.
	build.Env = os.Environ()
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}

	t.Run("put_preserves_positional_text_and_streams", func(t *testing.T) {
		root := t.TempDir()
		for _, text := range []string{"first", "-text", ""} {
			code, stdout, stderr := execute(t, binary, "", []string{"put", root, "alpha", text})
			if code != 0 || stdout != "" || stderr != "" {
				t.Fatalf("put = %d, stdout %q, stderr %q", code, stdout, stderr)
			}
			if got, err := os.ReadFile(filepath.Join(root, "alpha")); err != nil || string(got) != text {
				t.Fatalf("stored = %q, %v", got, err)
			}
		}
		code, stdout, stderr := execute(t, binary, "", []string{"put", root, "Alpha", "bad"})
		if code != 2 || stdout != "" || stderr == "" {
			t.Fatalf("invalid put = %d, stdout %q, stderr %q", code, stdout, stderr)
		}
		if got, err := os.ReadFile(filepath.Join(root, "alpha")); err != nil || len(got) != 0 {
			t.Fatalf("prior = %q, %v", got, err)
		}
	})
	t.Run("apply_success_and_retained_prefix", func(t *testing.T) {
		root := t.TempDir()
		code, stdout, stderr := execute(t, binary, "alpha,first\nalpha,second\nbeta,\" multi\nline \"\n", []string{"apply", "--dir", root})
		if code != 0 || stdout != "" || stderr != "" {
			t.Fatalf("apply = %d, %q, %q", code, stdout, stderr)
		}
		checkFile(t, filepath.Join(root, "alpha"), "second")
		checkFile(t, filepath.Join(root, "beta"), " multi\nline ")
		code, stdout, stderr = execute(t, binary, "alpha,accepted\nalpha, \t\ngamma,later\n", []string{"apply", "--dir", root})
		if code != 2 || stdout != "" || stderr == "" {
			t.Fatalf("rejected apply = %d, %q, %q", code, stdout, stderr)
		}
		checkFile(t, filepath.Join(root, "alpha"), "accepted")
		if _, err := os.Stat(filepath.Join(root, "gamma")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("later row = %v", err)
		}
	})
	t.Run("startup_rejections", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "snapshot")
		cases := []struct {
			name string
			args []string
		}{
			{"usage", nil},
			{"unknown", []string{"missing"}},
			{"put_usage", []string{"put", "directory"}},
			{"apply_missing_dir", []string{"apply"}},
			{"apply_extra", []string{"apply", "--dir", t.TempDir(), "extra"}},
			{"watch_scheme", []string{"watch", "--url", "ftp://example.test", "--file", path, "--interval", "1s"}},
			{"watch_host", []string{"watch", "--url", "http:/missing", "--file", path, "--interval", "1s"}},
			{"watch_port_without_host", []string{"watch", "--url", "http://:80", "--file", path, "--interval", "1s"}},
			{"watch_missing_file", []string{"watch", "--url", "http://example.test", "--interval", "1s"}},
			{"watch_zero", []string{"watch", "--url", "http://example.test", "--file", path, "--interval", "0s"}},
			{"watch_negative", []string{"watch", "--url", "http://example.test", "--file", path, "--interval", "-1s"}},
			{"watch_bad_duration", []string{"watch", "--url", "http://example.test", "--file", path, "--interval", "invalid"}},
			{"watch_unknown_flag", []string{"watch", "--unknown"}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				code, stdout, stderr := execute(t, binary, "", tc.args)
				if code != 2 || stdout != "" || stderr == "" {
					t.Fatalf("startup = %d, %q, %q", code, stdout, stderr)
				}
				if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("invalid startup wrote destination: %v", err)
				}
			})
		}
	})
	t.Run("failed_first_refresh", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
		}))
		t.Cleanup(server.Close)
		path := filepath.Join(t.TempDir(), "snapshot")
		if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
			t.Fatal(err)
		}
		code, stdout, stderr := execute(t, binary, "", []string{"watch", "--url", server.URL, "--file", path, "--interval", "20ms"})
		if code != 2 || stdout != "" || !strings.Contains(stderr, "503") {
			t.Fatalf("first failure = %d, %q, %q", code, stdout, stderr)
		}
		checkFile(t, path, "prior")
	})
	for _, signal := range []os.Signal{os.Interrupt, syscall.SIGTERM} {
		t.Run("watch_"+signal.String(), func(t *testing.T) {
			if runtime.GOOS == "windows" {
				t.Skip("POSIX signal process contract")
			}
			testWatchSignal(t, binary, signal)
		})
	}
}

func applicationEnv() []string {
	var env []string
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "SYSTEMROOT"} {
		if value, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+value)
		}
	}
	return env
}

func execute(t *testing.T, binary, input string, args []string) (int, string, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, args...)
	command.Env = applicationEnv()
	command.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("command %v exceeded deadline: %v", args, ctx.Err())
	}
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatalf("command %v: %v", args, err)
		}
		code = exit.ExitCode()
	}
	return code, stdout.String(), stderr.String()
}

func checkFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("%s = %q, %v; want %q", path, got, err, want)
	}
}

func testWatchSignal(t *testing.T, binary string, signal os.Signal) {
	var calls atomic.Int32
	thirdStarted := make(chan struct{})
	requestCanceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 3 {
			close(thirdStarted)
			<-r.Context().Done()
			close(requestCanceled)
			return
		}
		io.WriteString(w, fmt.Sprintf(`[{"key":"alpha","text":"round%d"}]`, n))
	}))
	path := filepath.Join(t.TempDir(), "snapshot")
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	command := exec.CommandContext(ctx, binary, "watch", "--url", server.URL, "--file", path, "--interval", "30ms")
	command.Env = applicationEnv()
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	joined := make(chan struct{})
	result := make(chan error, 1)
	if err := command.Start(); err != nil {
		cancel()
		server.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cancel()
		select {
		case <-joined:
		case <-time.After(3 * time.Second):
			t.Error("watch process did not join")
		}
		server.Close()
	})
	go func() { defer close(joined); result <- command.Wait() }()
	select {
	case <-thirdStarted:
	case err := <-result:
		t.Fatalf("watch exited before recurrence: %v, stderr %q", err, stderr.String())
	case <-time.After(4 * time.Second):
		t.Fatal("watch did not reach third request")
	}
	// Request three begins only after the previous publication has succeeded.
	checkFile(t, path, "[{\"key\":\"alpha\",\"text\":\"round2\"}]\n")
	if err := command.Process.Signal(signal); err != nil {
		t.Fatal(err)
	}
	select {
	case <-requestCanceled:
	case <-time.After(3 * time.Second):
		t.Fatal("in-flight HTTP request did not observe signal cancellation")
	}
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("watch signal exit: %v, stderr %q", err, stderr.String())
		}
	case <-time.After(3 * time.Second):
		t.Fatal("watch did not exit after signal")
	}
	if stdout.String() != "" || stderr.String() != "" {
		t.Fatalf("watch streams: stdout %q, stderr %q", stdout.String(), stderr.String())
	}
	if calls.Load() != 3 {
		t.Fatalf("requests = %d; want 3", calls.Load())
	}
	checkFile(t, path, "[{\"key\":\"alpha\",\"text\":\"round2\"}]\n")
}
