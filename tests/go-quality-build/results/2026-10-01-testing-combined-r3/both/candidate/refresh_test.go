package indexer_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"example.com/indexer"
)

var _ func(context.Context, *http.Client, string, string) error = indexer.Refresh

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type observedBody struct {
	io.Reader
	closes int
}

func (b *observedBody) Close() error { b.closes++; return nil }

func TestRefreshPublishesSnapshot(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"ordered records", `[ {"key":"beta","text":" line \n "}, {"key":"alpha","text":"one"}, {"key":"alpha","text":"two"} ] `, `[{"key":"beta","text":" line \n "},{"key":"alpha","text":"one"},{"key":"alpha","text":"two"}]` + "\n"},
		{"empty array", " [ ] \n\t", "[]\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path, old := seededSnapshot(t)
			reader, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { reader.Close() })
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			t.Cleanup(cancel)
			body := &observedBody{Reader: strings.NewReader(tc.input)}
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != http.MethodGet || r.URL.String() != "http://example.test/records" || r.Context() != ctx {
					t.Errorf("request = %s %s, context forwarded = %v", r.Method, r.URL, r.Context() == ctx)
				}
				return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}, nil
			})}
			if err := indexer.Refresh(ctx, client, "http://example.test/records", path); err != nil {
				t.Fatal(err)
			}
			assertFile(t, path, tc.want)
			gotOld, err := io.ReadAll(reader)
			if err != nil || string(gotOld) != old {
				t.Fatalf("open reader = %q, %v; want old snapshot %q", gotOld, err, old)
			}
			if calls != 1 || body.closes != 1 {
				t.Fatalf("requests = %d, body closes = %d; want 1 each", calls, body.closes)
			}
		})
	}
}

func TestRefreshRejectionsRetainSnapshot(t *testing.T) {
	readErr := errors.New("response read failed")
	for _, tc := range []struct {
		name, input string
		status      int
		invalid     bool
		readErr     error
	}{
		{"partial success status", `[{"key":"alpha","text":"new"}]`, 206, false, nil},
		{"server status", `[{"key":"alpha","text":"new"}]`, 500, false, nil},
		{"malformed", `[{`, 200, false, nil},
		{"object", `{"key":"alpha","text":"new"}`, 200, false, nil},
		{"null", `null`, 200, false, nil},
		{"second JSON", `[{"key":"alpha","text":"new"}] []`, 200, false, nil},
		{"trailing junk", `[{"key":"alpha","text":"new"}] junk`, 200, false, nil},
		{"bad key", `[{"key":"Alpha","text":"new"}]`, 200, true, nil},
		{"blank", `[{"key":"alpha","text":" \t\n\u2003"}]`, 200, true, nil},
		{"invalid later record", `[{"key":"alpha","text":"new"},{"key":"beta","text":""}]`, 200, true, nil},
		{"read after valid data", `[{"key":"alpha","text":"new"}]`, 200, false, readErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path, old := seededSnapshot(t)
			var reader io.Reader = strings.NewReader(tc.input)
			if tc.readErr != nil {
				reader = io.MultiReader(reader, failingReader{tc.readErr})
			}
			body := &observedBody{Reader: reader}
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: body, Header: make(http.Header)}, nil
			})}
			err := indexer.Refresh(context.Background(), client, "http://example.test/records", path)
			if err == nil {
				t.Fatal("accepted rejected response")
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("error = %v; want ErrInvalidRecord", err)
			}
			if tc.readErr != nil && !errors.Is(err, tc.readErr) {
				t.Fatalf("error = %v; want response read failure", err)
			}
			assertFile(t, path, old)
			if body.closes != 1 {
				t.Errorf("body closes = %d; want 1", body.closes)
			}
		})
	}
}

func TestRefreshFetchAndPublicationFailures(t *testing.T) {
	t.Run("fetch", func(t *testing.T) {
		path, old := seededSnapshot(t)
		wantErr := errors.New("transport failed")
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, wantErr
		})}
		if err := indexer.Refresh(context.Background(), client, "http://example.test/records", path); !errors.Is(err, wantErr) {
			t.Fatalf("error = %v; want transport failure", err)
		}
		assertFile(t, path, old)
	})
	t.Run("rename", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "snapshot")
		if err := os.Mkdir(path, 0700); err != nil {
			t.Fatal(err)
		}
		marker := filepath.Join(path, "marker")
		if err := os.WriteFile(marker, []byte("old"), 0600); err != nil {
			t.Fatal(err)
		}
		body := &observedBody{Reader: strings.NewReader(`[{"key":"alpha","text":"new"}]`)}
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}, nil
		})}
		if err := indexer.Refresh(context.Background(), client, "http://example.test/records", path); err == nil {
			t.Fatal("replaced directory")
		}
		assertFile(t, marker, "old")
		if body.closes != 1 {
			t.Errorf("body closes = %d; want 1", body.closes)
		}
		entries, err := os.ReadDir(root)
		if err != nil || len(entries) != 1 || entries[0].Name() != "snapshot" {
			t.Fatalf("publication left files = %v, %v", entries, err)
		}
	})
}

func TestRefreshRejectsRedirectWithoutChangingClient(t *testing.T) {
	var starts, targets, redirectPolicyCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			starts.Add(1)
			w.Header().Set("Location", "/target")
			w.WriteHeader(http.StatusFound)
		} else {
			targets.Add(1)
		}
		io.WriteString(w, `[{"key":"alpha","text":"new"}]`)
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		redirectPolicyCalls.Add(1)
		return nil
	}
	t.Cleanup(client.CloseIdleConnections)
	path, old := seededSnapshot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := indexer.Refresh(ctx, client, server.URL+"/start", path); err == nil {
		t.Fatal("accepted redirect")
	}
	assertFile(t, path, old)
	if starts.Load() != 1 || targets.Load() != 0 || redirectPolicyCalls.Load() != 0 {
		t.Fatalf("starts = %d, targets = %d, caller redirect calls = %d", starts.Load(), targets.Load(), redirectPolicyCalls.Load())
	}
	if err := client.CheckRedirect(nil, nil); err != nil || redirectPolicyCalls.Load() != 1 {
		t.Fatalf("caller policy changed: %v, calls %d", err, redirectPolicyCalls.Load())
	}
}

func TestRefreshNetworkCancellation(t *testing.T) {
	started, handlerDone := make(chan struct{}), make(chan struct{})
	handlerRelease := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-r.Context().Done():
		case <-handlerRelease:
		}
		close(handlerDone)
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	t.Cleanup(client.CloseIdleConnections)
	ctx, cancel := context.WithCancel(context.Background())
	path, old := seededSnapshot(t)
	result := make(chan error, 1)
	finished := make(chan struct{})
	t.Cleanup(func() {
		// Rescue the fixture even if Refresh regresses to ignoring cancellation.
		close(handlerRelease)
		cancel()
		waitClosed(t, finished, "Refresh cleanup")
	})
	go func() {
		defer close(finished)
		result <- indexer.Refresh(ctx, client, server.URL, path)
	}()
	waitClosed(t, started, "HTTP request")
	cancel()
	waitClosed(t, finished, "Refresh cancellation")
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v; want context.Canceled", err)
	}
	waitClosed(t, handlerDone, "HTTP handler cancellation")
	assertFile(t, path, old)
}

func seededSnapshot(t *testing.T) (string, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "snapshot.json")
	old := "old bytes\n\x00end"
	if err := os.WriteFile(path, []byte(old), 0600); err != nil {
		t.Fatal(err)
	}
	return path, old
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file %q = %q, %v; want %q", path, got, err, want)
	}
}

func waitClosed(t *testing.T, done <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}
