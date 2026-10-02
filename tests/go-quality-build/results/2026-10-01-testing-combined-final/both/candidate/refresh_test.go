package indexer_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"example.com/indexer"
)

var _ func(context.Context, *http.Client, string, string) error = indexer.Refresh

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type observedBody struct {
	io.Reader
	closes int
}

func (b *observedBody) Close() error { b.closes++; return nil }

func TestRefreshSuccessAndRejections(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		body, want string
		invalid    bool
	}{
		{"ordered original text", 200, ` [ {"key":"beta","text":"  second  "}, {"key":"alpha","text":"first"}, {"key":"beta","text":"last"} ] `, "[{\"key\":\"beta\",\"text\":\"  second  \"},{\"key\":\"alpha\",\"text\":\"first\"},{\"key\":\"beta\",\"text\":\"last\"}]\n", false},
		{"empty array", 200, "[]\n\t ", "[]\n", false},
		{"partial content is rejected", 206, `[{"key":"alpha","text":"valid"}]`, "", false},
		{"no content", 204, `[]`, "", false},
		{"server error", 503, `[]`, "", false},
		{"null is not array", 200, `null`, "", false},
		{"object is not array", 200, `{"key":"alpha","text":"valid"}`, "", false},
		{"malformed", 200, `[{`, "", false},
		{"trailing array", 200, `[][]`, "", false},
		{"trailing garbage", 200, `[] garbage`, "", false},
		{"invalid second key", 200, `[{"key":"alpha","text":"valid"},{"key":"A","text":"valid"}]`, "", true},
		{"missing key", 200, `[{"text":"valid"}]`, "", true},
		{"blank text", 200, `[{"key":"alpha","text":" \t\n\u2003"}]`, "", true},
		{"missing text", 200, `[{"key":"alpha"}]`, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "snapshot")
			prior := []byte("prior bytes\x00\n")
			if err := os.WriteFile(path, prior, 0600); err != nil {
				t.Fatal(err)
			}
			body := &observedBody{Reader: strings.NewReader(tc.body)}
			ctx := context.WithValue(context.Background(), contextKey{}, "caller")
			calls := 0
			client := &http.Client{Transport: transportFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				if req.Method != "GET" || req.URL.String() != "http://fixture.invalid/records" || req.Context() != ctx {
					return nil, errors.New("request did not preserve method, endpoint or context")
				}
				return &http.Response{StatusCode: tc.status, Body: body, Header: make(http.Header), Request: req}, nil
			})}
			err := indexer.Refresh(ctx, client, "http://fixture.invalid/records", path)
			if tc.want != "" {
				if err != nil {
					t.Fatalf("Refresh: %v", err)
				}
			} else if err == nil {
				t.Fatal("rejected response succeeded")
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Errorf("Refresh error = %v; want validation identity", err)
			}
			expected := string(prior)
			if tc.want != "" {
				expected = tc.want
			}
			assertFile(t, path, expected)
			if calls != 1 || body.closes != 1 {
				t.Errorf("requests=%d closes=%d; want 1 each", calls, body.closes)
			}
		})
	}
}

type contextKey struct{}

func TestRefreshFailureStages(t *testing.T) {
	readErr, fetchErr := errors.New("body read failed"), errors.New("fetch failed")
	for _, stage := range []string{"fetch", "read", "create", "rename"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			path := filepath.Join(root, "snapshot")
			if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
				t.Fatal(err)
			}
			target := path
			if stage == "create" {
				target = filepath.Join(root, "missing", "snapshot")
			}
			if stage == "rename" {
				target = filepath.Join(root, "directory")
				if err := os.Mkdir(target, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(target, "prior"), []byte("directory bytes"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			var reader io.Reader = strings.NewReader("[]")
			if stage == "read" {
				reader = &failingReader{strings.NewReader("[]"), readErr}
			}
			body := &observedBody{Reader: reader}
			client := &http.Client{Transport: transportFunc(func(req *http.Request) (*http.Response, error) {
				if stage == "fetch" {
					return nil, fetchErr
				}
				return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header), Request: req}, nil
			})}
			err := indexer.Refresh(context.Background(), client, "http://fixture.invalid", target)
			if err == nil {
				t.Fatal("failure stage succeeded")
			}
			if stage == "fetch" && !errors.Is(err, fetchErr) {
				t.Errorf("fetch cause lost: %v", err)
			}
			if stage == "read" && !errors.Is(err, readErr) {
				t.Errorf("read cause lost: %v", err)
			}
			expectedCloses := 1
			if stage == "fetch" {
				expectedCloses = 0
			}
			if body.closes != expectedCloses {
				t.Errorf("body closes=%d; want %d", body.closes, expectedCloses)
			}
			assertFile(t, path, "prior")
			if stage == "rename" {
				assertFile(t, filepath.Join(target, "prior"), "directory bytes")
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), ".record-") {
					t.Errorf("temporary file remains: %s", entry.Name())
				}
			}
		})
	}
}

func TestRefreshPOSIXSnapshot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX replacement semantics; Windows is outside the execution claim")
	}
	t.Parallel()
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("complete old snapshot\n"), 0600); err != nil {
		t.Fatal(err)
	}
	old, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { old.Close() })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/records" {
			http.Error(w, "wrong request", 400)
			return
		}
		io.WriteString(w, `[{"key":"alpha","text":"new"}]`)
	}))
	t.Cleanup(server.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := indexer.Refresh(ctx, server.Client(), server.URL+"/records", path); err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(old)
	if err != nil || string(got) != "complete old snapshot\n" {
		t.Fatalf("old reader = %q, %v", got, err)
	}
	assertFile(t, path, "[{\"key\":\"alpha\",\"text\":\"new\"}]\n")
}

func TestRefreshRejectsRedirectWithoutChangingClient(t *testing.T) {
	t.Parallel()
	var requests, policyCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/records", http.StatusFound)
			return
		}
		io.WriteString(w, "[]")
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	client.Timeout = 5 * time.Second
	client.CheckRedirect = func(*http.Request, []*http.Request) error { policyCalls.Add(1); return nil }
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := indexer.Refresh(ctx, client, server.URL+"/redirect", path); err == nil {
		t.Fatal("redirect succeeded")
	}
	if requests.Load() != 1 || policyCalls.Load() != 0 {
		t.Fatalf("requests=%d policy calls=%d", requests.Load(), policyCalls.Load())
	}
	assertFile(t, path, "old")
	// The borrowed client's original redirect behavior remains available to its owner.
	response, err := client.Get(server.URL + "/redirect")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if policyCalls.Load() != 1 || requests.Load() != 3 {
		t.Errorf("client policy changed: requests=%d policy=%d", requests.Load(), policyCalls.Load())
	}
}

func TestRefreshTransportCancellation(t *testing.T) {
	t.Parallel()
	started, handlerDone := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		close(handlerDone)
	}))
	t.Cleanup(server.Close)
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	done := make(chan error, 1)
	exited := make(chan struct{})
	t.Cleanup(func() {
		cancel()
		select {
		case <-exited:
		case <-time.After(5 * time.Second):
			t.Error("Refresh did not stop during cleanup")
		}
	})
	go func() { done <- indexer.Refresh(ctx, server.Client(), server.URL, path); close(exited) }()
	await(t, started, "request startup")
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Refresh = %v; want cancellation", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Refresh did not stop")
	}
	await(t, handlerDone, "transport cancellation")
	assertFile(t, path, "old")
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file %s = %q, %v; want %q", path, got, err, want)
	}
}
func await(t *testing.T, ch <-chan struct{}, event string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", event)
	}
}
