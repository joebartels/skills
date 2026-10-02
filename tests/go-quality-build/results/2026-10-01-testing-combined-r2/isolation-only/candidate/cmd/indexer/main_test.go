package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func TestExistingPut(t *testing.T) {
	if err := run([]string{"put", t.TempDir(), "alpha", "one"}); err != nil {
		t.Fatal(err)
	}
	if err := run(nil); err == nil {
		t.Fatal("missing usage rejection")
	}
}

// Child processes receive only these explicit, non-secret environment values.
// In particular, caller proxy variables cannot reroute the local HTTP fixture.
func childEnvironment() []string {
	env := []string{"GOCACHE=/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN=local"}
	for _, name := range []string{"PATH", "TMPDIR", "SYSTEMROOT"} {
		if value, ok := os.LookupEnv(name); ok {
			env = append(env, name+"="+value)
		}
	}
	return env
}

func buildCommand(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "indexer")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	cmd.Env = childEnvironment()
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build actual command: %v\n%s", err, output)
	}
	return binary
}

type outcome struct {
	code           int
	stdout, stderr string
}

func invoke(t *testing.T, binary, dir, stdin string, args ...string) outcome {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir, cmd.Env = dir, childEnvironment()
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("command deadline: %v, stderr=%q", ctx.Err(), stderr.String())
	}
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("start command: %v", err)
		}
		code = exitErr.ExitCode()
	}
	return outcome{code, stdout.String(), stderr.String()}
}

func assertOutcome(t *testing.T, got outcome, code int) {
	t.Helper()
	if got.code != code || got.stdout != "" || (code == 0 && got.stderr != "") || (code != 0 && got.stderr == "") {
		t.Fatalf("process code=%d stdout=%q stderr=%q; want code=%d, no stdout and matching diagnostic convention", got.code, got.stdout, got.stderr, code)
	}
}

func assertBytes(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("ReadFile(%q)=%q, %v; want %q", path, got, err, want)
	}
}

func TestCommandPutAndApplyProcesses(t *testing.T) {
	binary := buildCommand(t)
	t.Run("legacy_put_dash_arguments", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		assertOutcome(t, invoke(t, binary, dir, "", "put", "-index", "alpha", "-literal text"), 0)
		assertBytes(t, filepath.Join(dir, "-index", "alpha"), "-literal text")
		assertOutcome(t, invoke(t, binary, dir, "", "put", "-index", "Alpha", "bad"), 2)
		assertBytes(t, filepath.Join(dir, "-index", "alpha"), "-literal text")
	})
	t.Run("apply_success", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		assertOutcome(t, invoke(t, binary, root, "alpha,one\nalpha,two\nbeta,\"comma, text\"\n", "apply", "--dir", root), 0)
		assertBytes(t, filepath.Join(root, "alpha"), "two")
		assertBytes(t, filepath.Join(root, "beta"), "comma, text")
	})
	t.Run("apply_accepted_prefix", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		assertOutcome(t, invoke(t, binary, root, "alpha,accepted\nalpha, \t\ngamma,later\n", "apply", "--dir", root), 2)
		assertBytes(t, filepath.Join(root, "alpha"), "accepted")
		if _, err := os.Stat(filepath.Join(root, "gamma")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("later row = %v", err)
		}
	})
	t.Run("apply_parse_failure", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		assertOutcome(t, invoke(t, binary, root, "alpha,accepted\nbeta,\"bad\n", "apply", "--dir", root), 2)
		assertBytes(t, filepath.Join(root, "alpha"), "accepted")
	})
	t.Run("put_filesystem_failure", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		blocked := filepath.Join(root, "file")
		if err := os.WriteFile(blocked, []byte("prior"), 0600); err != nil {
			t.Fatal(err)
		}
		assertOutcome(t, invoke(t, binary, root, "", "put", blocked, "alpha", "text"), 2)
		assertBytes(t, blocked, "prior")
	})
}

func TestCommandInvalidStartupProcesses(t *testing.T) {
	binary := buildCommand(t)
	cases := []struct {
		name string
		args []string
	}{
		{"missing_command", nil},
		{"unknown_command", []string{"unknown"}},
		{"put_usage", []string{"put", "dir", "key"}},
		{"apply_missing_dir", []string{"apply"}},
		{"apply_empty_dir", []string{"apply", "--dir", ""}},
		{"apply_unknown_flag", []string{"apply", "--unknown", "value"}},
		{"apply_extra_argument", []string{"apply", "--dir", "dir", "extra"}},
		{"watch_missing_options", []string{"watch"}},
		{"watch_relative_url", []string{"watch", "--url", "/relative", "--file", "file", "--interval", "1ms"}},
		{"watch_ftp", []string{"watch", "--url", "ftp://host/file", "--file", "file", "--interval", "1ms"}},
		{"watch_no_host", []string{"watch", "--url", "http:///file", "--file", "file", "--interval", "1ms"}},
		{"watch_empty_hostname", []string{"watch", "--url", "http://:80/file", "--file", "file", "--interval", "1ms"}},
		{"watch_bad_url", []string{"watch", "--url", "http://%", "--file", "file", "--interval", "1ms"}},
		{"watch_empty_path", []string{"watch", "--url", "http://host", "--file", "", "--interval", "1ms"}},
		{"watch_zero", []string{"watch", "--url", "http://host", "--file", "file", "--interval", "0"}},
		{"watch_negative", []string{"watch", "--url", "http://host", "--file", "file", "--interval", "-1s"}},
		{"watch_bad_duration", []string{"watch", "--url", "http://host", "--file", "file", "--interval", "soon"}},
		{"watch_unknown_flag", []string{"watch", "--unknown"}},
		{"watch_extra_argument", []string{"watch", "--url", "http://host", "--file", "file", "--interval", "1ms", "extra"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertOutcome(t, invoke(t, binary, t.TempDir(), "", tc.args...), 2)
		})
	}
}

func TestWatchFirstRefreshFailureProcesses(t *testing.T) {
	binary := buildCommand(t)
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"status", 500, "[]"},
		{"malformed", 200, "["},
		{"validation", 200, `[{"key":"alpha","text":""}]`},
	} {
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
			if err := os.WriteFile(path, []byte("prior\x00\n"), 0600); err != nil {
				t.Fatal(err)
			}
			assertOutcome(t, invoke(t, binary, root, "", "watch", "--url", server.URL, "--file", path, "--interval", "10ms"), 2)
			if calls.Load() != 1 {
				t.Fatalf("request count = %d", calls.Load())
			}
			assertBytes(t, path, "prior\x00\n")
		})
	}
}

func waitEvent(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-ch:
	case <-timer.C:
		t.Fatalf("timed out waiting for %s", what)
	}
}

func TestWatchRecurrenceAndSignalProcesses(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX signal contract")
	}
	binary := buildCommand(t)
	for _, tc := range []struct {
		name   string
		signal os.Signal
	}{
		{"interrupt", os.Interrupt},
		{"termination", syscall.SIGTERM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			third, canceled := make(chan struct{}), make(chan struct{})
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := calls.Add(1)
				if n == 3 {
					close(third)
					<-r.Context().Done()
					close(canceled)
					return
				}
				text := "first"
				if n == 2 {
					text = "second"
				}
				json.NewEncoder(w).Encode([]struct {
					Key  string `json:"key"`
					Text string `json:"text"`
				}{{"alpha", text}})
			}))
			t.Cleanup(server.Close)
			root := t.TempDir()
			path := filepath.Join(root, "snapshot")
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "watch", "--url", server.URL, "--file", path, "--interval", "20ms")
			cmd.Dir, cmd.Env = root, childEnvironment()
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			finished := make(chan struct{})
			var waitErr error
			go func() { waitErr = cmd.Wait(); close(finished) }()
			t.Cleanup(func() {
				cancel()
				cmd.Process.Kill()
				waitEvent(t, finished, "watch process cleanup join")
			})
			select {
			case <-third:
			case <-finished:
				t.Fatalf("watch exited before recurrence: %v, stdout=%q, stderr=%q", waitErr, stdout.String(), stderr.String())
			case <-ctx.Done():
				t.Fatal("watch never reached third request")
			}
			// Third request startup proves two complete successful publications.
			assertBytes(t, path, "[{\"key\":\"alpha\",\"text\":\"second\"}]\n")
			if err := cmd.Process.Signal(tc.signal); err != nil {
				t.Fatal(err)
			}
			waitEvent(t, canceled, "in-flight HTTP cancellation")
			waitEvent(t, finished, "watch process signal join")
			if waitErr != nil || stdout.Len() != 0 || stderr.Len() != 0 {
				t.Fatalf("watch result=%v stdout=%q stderr=%q", waitErr, stdout.String(), stderr.String())
			}
			if calls.Load() != 3 {
				t.Fatalf("request count=%d; want 3", calls.Load())
			}
			assertBytes(t, path, "[{\"key\":\"alpha\",\"text\":\"second\"}]\n")
		})
	}
}
