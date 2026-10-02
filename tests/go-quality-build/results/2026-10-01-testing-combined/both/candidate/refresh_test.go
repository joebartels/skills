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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackedBody struct {
	io.Reader
	closes int
}

func (b *trackedBody) Close() error { b.closes++; return nil }

func TestRefreshPublishesExactRepresentationAndClosesBody(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"ordered original text", " [ {\"key\":\"beta\",\"text\":\"  first  \"}, {\"key\":\"alpha\",\"text\":\"second\"}, {\"key\":\"beta\",\"text\":\"last\"} ] \n\t", "[{\"key\":\"beta\",\"text\":\"  first  \"},{\"key\":\"alpha\",\"text\":\"second\"},{\"key\":\"beta\",\"text\":\"last\"}]\n"},
		{"empty array", "[]", "[]\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "records.json")
			body := &trackedBody{Reader: strings.NewReader(tc.input)}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != "GET" || r.URL.String() != "http://records.example/list?q=one" || r.Context() != ctx {
					t.Errorf("request = %s %s, context forwarded = %v", r.Method, r.URL, r.Context() == ctx)
				}
				return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}, nil
			})}
			if err := indexer.Refresh(ctx, client, "http://records.example/list?q=one", path); err != nil {
				t.Fatal(err)
			}
			assertFile(t, path, tc.want)
			if calls != 1 || body.closes != 1 {
				t.Fatalf("requests = %d, body closes = %d; want one each", calls, body.closes)
			}
		})
	}
}

func TestRefreshRejectionsRetainBytesAndCloseBody(t *testing.T) {
	readErr := errors.New("body read failed")
	for _, tc := range []struct {
		name   string
		status int
		input  io.Reader
		cause  error
	}{
		{"status 201", 201, strings.NewReader("[]"), nil},
		{"status 500", 500, strings.NewReader("[]"), nil},
		{"syntax", 200, strings.NewReader("["), nil},
		{"object", 200, strings.NewReader("{}"), nil},
		{"null", 200, strings.NewReader("null"), indexer.ErrInvalidRecord},
		{"second array", 200, strings.NewReader("[][]"), nil},
		{"trailing garbage", 200, strings.NewReader("[] trailing"), nil},
		{"invalid key", 200, strings.NewReader("[{\"key\":\"Alpha\",\"text\":\"one\"}]"), indexer.ErrInvalidRecord},
		{"blank text", 200, strings.NewReader("[{\"key\":\"alpha\",\"text\":\" \\t\\u2003\"}]"), indexer.ErrInvalidRecord},
		{"later invalid record", 200, strings.NewReader("[{\"key\":\"alpha\",\"text\":\"first\"},{\"key\":\"beta\",\"text\":\"\"}]"), indexer.ErrInvalidRecord},
		{"read before JSON", 200, failingReader{readErr}, readErr},
		{"read after array", 200, io.MultiReader(strings.NewReader("[]"), failingReader{readErr}), readErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "records.json")
			prior := "previous bytes\x00\n  "
			writeFile(t, path, prior)
			body := &trackedBody{Reader: tc.input}
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: body, Header: make(http.Header)}, nil
			})}
			err := indexer.Refresh(context.Background(), client, "http://records.example", path)
			if err == nil || (tc.cause != nil && !errors.Is(err, tc.cause)) {
				t.Fatalf("Refresh = %v; want rejection with cause %v", err, tc.cause)
			}
			assertFile(t, path, prior)
			if body.closes != 1 {
				t.Fatalf("body closes = %d; want 1", body.closes)
			}
		})
	}
}

func TestRefreshFetchFailureRetainsBytes(t *testing.T) {
	wantErr := errors.New("transport failed")
	path := filepath.Join(t.TempDir(), "records.json")
	writeFile(t, path, "prior")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, wantErr })}
	if err := indexer.Refresh(context.Background(), client, "http://records.example", path); !errors.Is(err, wantErr) {
		t.Fatalf("Refresh = %v; want transport cause", err)
	}
	assertFile(t, path, "prior")
}

func TestRefreshDoesNotFollowRedirectOrMutateClient(t *testing.T) {
	path := filepath.Join(t.TempDir(), "records.json")
	writeFile(t, path, "prior")
	body := &trackedBody{Reader: strings.NewReader("redirect body")}
	calls, redirects := 0, 0
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { redirects++; return nil },
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if calls != 1 {
				t.Error("redirect issued another GET")
			}
			return &http.Response{StatusCode: 302, Body: body, Header: http.Header{"Location": {"http://records.example/redirected"}}}, nil
		}),
	}
	if err := indexer.Refresh(context.Background(), client, "http://records.example", path); err == nil {
		t.Fatal("redirect accepted")
	}
	assertFile(t, path, "prior")
	if calls != 1 || body.closes != 1 || redirects != 0 {
		t.Fatalf("requests %d, closes %d, redirect policy calls %d", calls, body.closes, redirects)
	}
	if err := client.CheckRedirect(nil, nil); err != nil || redirects != 1 {
		t.Fatal("borrowed client's redirect policy changed")
	}
}

func TestRefreshPublicationFailureClosesBody(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "destination")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(path, "prior"), "retained")
	body := &trackedBody{Reader: strings.NewReader("[]")}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}, nil
	})}
	if err := indexer.Refresh(context.Background(), client, "http://records.example", path); err == nil {
		t.Fatal("publishing over directory succeeded")
	}
	assertFile(t, filepath.Join(path, "prior"), "retained")
	if body.closes != 1 {
		t.Fatalf("body closes = %d; want 1", body.closes)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 {
		t.Fatalf("publication left temporary files: %v, %v", entries, err)
	}
}

func TestRefreshPOSIXOpenSnapshot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("open-file replacement semantics are a POSIX contract")
	}
	path := filepath.Join(t.TempDir(), "records.json")
	writeFile(t, path, "complete old snapshot")
	old, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { old.Close() })
	body := &trackedBody{Reader: strings.NewReader("[{\"key\":\"alpha\",\"text\":\"new\"}]")}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}, nil
	})}
	if err := indexer.Refresh(context.Background(), client, "http://records.example", path); err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(old)
	if err != nil || string(got) != "complete old snapshot" {
		t.Fatalf("old reader = %q, %v", got, err)
	}
	assertFile(t, path, "[{\"key\":\"alpha\",\"text\":\"new\"}]\n")
}

func TestRefreshRealHTTPCancellation(t *testing.T) {
	started := make(chan struct{})
	ended := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "[{\"key\":\"alpha\",\"text\":\"")
		w.(http.Flusher).Flush()
		close(started)
		<-r.Context().Done()
		close(ended)
	}))
	t.Cleanup(server.Close)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	joined := make(chan struct{})
	path := filepath.Join(t.TempDir(), "records.json")
	writeFile(t, path, "prior")
	t.Cleanup(func() { cancel(); awaitEvent(t, joined, "Refresh cleanup join") })
	go func() {
		done <- indexer.Refresh(ctx, server.Client(), server.URL, path)
		close(joined)
	}()
	awaitEvent(t, started, "HTTP request start")
	cancel()
	err := awaitResult(t, done)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Refresh = %v; want context cancellation", err)
	}
	awaitEvent(t, ended, "HTTP handler cancellation")
	assertFile(t, path, "prior")
}

func TestRefreshRealHTTP(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/records" {
			t.Errorf("request = %s %s", r.Method, r.URL)
		}
		io.WriteString(w, "[{\"key\":\"alpha\",\"text\":\"network\"}]")
	}))
	t.Cleanup(server.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "records.json")
	if err := indexer.Refresh(ctx, server.Client(), server.URL+"/records", path); err != nil {
		t.Fatal(err)
	}
	assertFile(t, path, "[{\"key\":\"alpha\",\"text\":\"network\"}]\n")
	if requests.Load() != 1 {
		t.Fatalf("requests = %d; want 1", requests.Load())
	}
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file %s = %q, %v; want %q", path, got, err, want)
	}
}
