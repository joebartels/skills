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
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestCommandProcessContracts(t *testing.T) {
	binary := buildIndexer(t)
	t.Run("put positional text", func(t *testing.T) {
		root := t.TempDir()
		assertProcess(t, binary, []string{"put", root, "alpha", "-legacy text"}, "", 0)
		assertContents(t, filepath.Join(root, "alpha"), "-legacy text")
		assertProcess(t, binary, []string{"put", root, "alpha", ""}, "", 0)
		assertContents(t, filepath.Join(root, "alpha"), "")
	})
	t.Run("put rejection preserves value", func(t *testing.T) {
		root := t.TempDir()
		assertProcess(t, binary, []string{"put", root, "alpha", "prior"}, "", 0)
		assertProcess(t, binary, []string{"put", root, "../alpha", "bad"}, "", 2)
		assertContents(t, filepath.Join(root, "alpha"), "prior")
	})
	t.Run("apply ordered stdin", func(t *testing.T) {
		root := t.TempDir()
		assertProcess(t, binary, []string{"apply", "--dir", root}, "alpha,first\nbeta,\"comma, text\"\nalpha,last\n", 0)
		assertContents(t, filepath.Join(root, "alpha"), "last")
		assertContents(t, filepath.Join(root, "beta"), "comma, text")
	})
	t.Run("apply write failure prefix", func(t *testing.T) {
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "beta"), 0700); err != nil {
			t.Fatal(err)
		}
		assertProcess(t, binary, []string{"apply", "--dir", root}, "alpha,accepted\nbeta,rejected\ngamma,later\n", 2)
		assertContents(t, filepath.Join(root, "alpha"), "accepted")
		if _, err := os.Stat(filepath.Join(root, "gamma")); !os.IsNotExist(err) {
			t.Fatalf("later row effect: %v", err)
		}
	})
	for _, tc := range []struct{ name, rejected string }{
		{"apply validation prefix", "alpha,\n"},
		{"apply parse prefix", "alpha,rejected,extra\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			assertProcess(t, binary, []string{"apply", "--dir", root}, "alpha,accepted\n"+tc.rejected+"gamma,later\n", 2)
			assertContents(t, filepath.Join(root, "alpha"), "accepted")
			if _, err := os.Stat(filepath.Join(root, "gamma")); !os.IsNotExist(err) {
				t.Fatalf("later row effect: %v", err)
			}
		})
	}
	for _, args := range [][]string{nil, {"put"}, {"apply"}, {"apply", "--unknown"}, {"apply", "--dir", "somewhere", "extra"}, {"unknown"}} {
		t.Run(fmt.Sprintf("usage %q", args), func(t *testing.T) { assertProcess(t, binary, args, "", 2) })
	}
}

func TestWatchProcessContracts(t *testing.T) {
	binary := buildIndexer(t)
	for _, tc := range []struct{ name, endpoint, path, interval string }{
		{"missing URL", "", "out", "1s"},
		{"missing host", "http:///path", "out", "1s"},
		{"unsupported scheme", "file://host/path", "out", "1s"},
		{"missing path", "http://127.0.0.1:1", "", "1s"},
		{"zero interval", "http://127.0.0.1:1", "out", "0"},
		{"negative interval", "http://127.0.0.1:1", "out", "-1s"},
		{"malformed duration", "http://127.0.0.1:1", "out", "fast"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertProcess(t, binary, []string{"watch", "--url", tc.endpoint, "--file", tc.path, "--interval", tc.interval}, "", 2)
		})
	}
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"first status failure", 503, "[]"},
		{"first validation failure", 200, "[{\"key\":\"alpha\",\"text\":\"\"}]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(tc.status)
				io.WriteString(w, tc.body)
			}))
			t.Cleanup(server.Close)
			path := filepath.Join(t.TempDir(), "records.json")
			if err := os.WriteFile(path, []byte("prior bytes"), 0600); err != nil {
				t.Fatal(err)
			}
			assertProcess(t, binary, []string{"watch", "--url", server.URL, "--file", path, "--interval", "10ms"}, "", 2)
			assertContents(t, path, "prior bytes")
			if calls.Load() != 1 {
				t.Fatalf("requests = %d; want failed first request only", calls.Load())
			}
		})
	}
	t.Run("recurrence and in flight termination", func(t *testing.T) {
		for _, termination := range []os.Signal{syscall.SIGTERM, os.Interrupt} {
			t.Run(termination.String(), func(t *testing.T) {
				requests := make(chan int32, 8)
				canceled := make(chan struct{})
				secondResponse := make(chan struct{})
				var calls atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					n := calls.Add(1)
					requests <- n
					if n == 2 {
						select {
						case <-secondResponse:
						case <-r.Context().Done():
							return
						}
					}
					if n >= 3 {
						<-r.Context().Done()
						if n == 3 {
							close(canceled)
						}
						return
					}
					fmt.Fprintf(w, "[{\"key\":\"alpha\",\"text\":\"cycle %d\"}]", n)
				}))
				t.Cleanup(server.Close)
				path := filepath.Join(t.TempDir(), "records.json")
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				cmd := exec.CommandContext(ctx, binary, "watch", "--url", server.URL, "--file", path, "--interval", "150ms")
				cmd.Env = childEnvironment()
				cmd.WaitDelay = time.Second
				var stdout, stderr bytes.Buffer
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				if err := cmd.Start(); err != nil {
					cancel()
					t.Fatal(err)
				}
				done := make(chan error, 1)
				joined := make(chan struct{})
				t.Cleanup(func() {
					cancel()
					select {
					case <-joined:
					case <-time.After(5 * time.Second):
						t.Error("watch process did not join during cleanup")
					}
				})
				go func() { done <- cmd.Wait(); close(joined) }()
				for n := int32(1); n <= 2; n++ {
					select {
					case got := <-requests:
						if got != n {
							t.Fatalf("request sequence = %d; want %d", got, n)
						}
					case <-joined:
						t.Fatalf("watch exited before refresh %d", n)
					case <-time.After(5 * time.Second):
						t.Fatalf("refresh %d did not start", n)
					}
					waitForContents(t, path, fmt.Sprintf("[{\"key\":\"alpha\",\"text\":\"cycle %d\"}]\n", n), joined)
					if n == 1 {
						close(secondResponse)
					}
				}
				select {
				case n := <-requests:
					if n != 3 {
						t.Fatalf("third request = %d", n)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("third request did not start")
				}
				if err := cmd.Process.Signal(termination); err != nil {
					t.Fatal(err)
				}
				select {
				case err := <-done:
					if err != nil {
						t.Fatalf("terminated watch = %v, stderr %q", err, stderr.String())
					}
				case <-time.After(5 * time.Second):
					t.Fatal("terminated watch did not join")
				}
				select {
				case <-canceled:
				case <-time.After(5 * time.Second):
					t.Fatal("in-flight HTTP request was not canceled")
				}
				if stdout.Len() != 0 || stderr.Len() != 0 {
					t.Fatalf("successful watch stdout %q, stderr %q; want empty", stdout.String(), stderr.String())
				}
				assertContents(t, path, "[{\"key\":\"alpha\",\"text\":\"cycle 2\"}]\n")
				if calls.Load() != 3 {
					t.Fatalf("requests after shutdown = %d; want 3", calls.Load())
				}
			})
		}
	})
}

func buildIndexer(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "indexer")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", path, ".")
	cmd.Env = childEnvironment()
	cmd.WaitDelay = time.Second
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build command: %v\n%s", err, output)
	}
	return path
}

func childEnvironment() []string {
	env := []string{"GOCACHE=/private/tmp/go-quality-testing-cache", "GOTOOLCHAIN=local"}
	for _, key := range []string{"PATH", "HOME", "TMPDIR"} {
		if value, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+value)
		}
	}
	return env
}

func assertProcess(t *testing.T, binary string, args []string, input string, wantCode int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = childEnvironment()
	cmd.WaitDelay = time.Second
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("command timed out: %v", ctx.Err())
	}
	code := 0
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	if code != wantCode || stdout.Len() != 0 || wantCode == 0 && stderr.Len() != 0 || wantCode != 0 && stderr.Len() == 0 {
		t.Fatalf("%v: status %d, stdout %q, stderr %q; want status %d and matching streams", args, code, stdout.String(), stderr.String(), wantCode)
	}
}

func assertContents(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file = %q, %v; want %q", got, err, want)
	}
}

func waitForContents(t *testing.T, path, want string, joined <-chan struct{}) {
	t.Helper()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		got, err := os.ReadFile(path)
		if err == nil && string(got) == want {
			return
		}
		select {
		case <-timer.C:
			t.Fatalf("publication = %q, %v; want %q", got, err, want)
		case <-joined:
			t.Fatal("watch exited before publication")
		case <-ticker.C:
		}
	}
}
