package indexer_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/indexer"
)

var _ func(context.Context, *http.Client, string, string) error = indexer.Refresh

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error { b.closed = true; return nil }

func TestRefreshRepresentationsAndSnapshots(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"ordered duplicates", " [ {\"key\":\"beta\",\"text\":\" first \"}, {\"key\":\"alpha\",\"text\":\"line\\nnext\"}, {\"key\":\"beta\",\"text\":\"last\"} ] \n", "[{\"key\":\"beta\",\"text\":\" first \"},{\"key\":\"alpha\",\"text\":\"line\\nnext\"},{\"key\":\"beta\",\"text\":\"last\"}]\n"},
		{"empty array", "[]", "[]\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "snapshot")
			const old = "old exact bytes\x00\n"
			if err := os.WriteFile(path, []byte(old), 0600); err != nil {
				t.Fatal(err)
			}
			before, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer before.Close()
			body := &trackedBody{Reader: strings.NewReader(tc.input)}
			ctx := context.WithValue(context.Background(), struct{}{}, "caller")
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Context() != ctx || r.Method != "GET" || r.URL.String() != "https://records.test/data" {
					t.Errorf("request = %#v", r)
				}
				return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}, nil
			})}
			if err := indexer.Refresh(ctx, client, "https://records.test/data", path); err != nil {
				t.Fatal(err)
			}
			if calls != 1 || !body.closed {
				t.Fatalf("calls = %d, body closed = %v", calls, body.closed)
			}
			assertFile(t, path, tc.want)
			got, err := io.ReadAll(before)
			if err != nil || string(got) != old {
				t.Fatalf("old reader = %q, %v", got, err)
			}
			entries, err := os.ReadDir(filepath.Dir(path))
			if err != nil || len(entries) != 1 {
				t.Fatalf("temporary files remain: %v, %v", entries, err)
			}
		})
	}
}

func TestRefreshRejectionRetainsDestinationAndClosesBody(t *testing.T) {
	readErr := errors.New("body interrupted")
	for _, tc := range []struct {
		name    string
		status  int
		body    io.Reader
		invalid bool
		cause   error
	}{
		{"partial content", 206, strings.NewReader(`[{"key":"alpha","text":"valid"}]`), false, nil},
		{"server error", 503, strings.NewReader(`[]`), false, nil},
		{"read failure after data", 200, io.MultiReader(strings.NewReader(`[]`), failingReader{readErr}), false, readErr},
		{"null", 200, strings.NewReader(`null`), false, nil},
		{"object", 200, strings.NewReader(`{"key":"alpha","text":"valid"}`), false, nil},
		{"broken JSON", 200, strings.NewReader(`[{`), false, nil},
		{"second array", 200, strings.NewReader(`[] []`), false, nil},
		{"trailing junk", 200, strings.NewReader(`[] x`), false, nil},
		{"invalid key", 200, strings.NewReader(`[{"key":"Alpha","text":"valid"}]`), true, nil},
		{"blank text", 200, strings.NewReader(`[{"key":"alpha","text":" \t\n"}]`), true, nil},
		{"invalid later record", 200, strings.NewReader(`[{"key":"alpha","text":"valid"},{"key":"beta","text":""}]`), true, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "snapshot")
			const old = "prior bytes \n\x00"
			if err := os.WriteFile(path, []byte(old), 0600); err != nil {
				t.Fatal(err)
			}
			body := &trackedBody{Reader: tc.body}
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: body}, nil
			})}
			err := indexer.Refresh(context.Background(), client, "https://records.test", path)
			if err == nil {
				t.Fatal("expected rejection")
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("validation identity = %v", err)
			}
			if tc.cause != nil && !errors.Is(err, tc.cause) {
				t.Fatalf("read identity = %v", err)
			}
			if !body.closed {
				t.Fatal("response body left open")
			}
			assertFile(t, path, old)
		})
	}
}

func TestRefreshFetchFailureAndContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	want := errors.New("fetch failed")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, want })}
	if err := indexer.Refresh(context.Background(), client, "https://records.test", path); !errors.Is(err, want) {
		t.Fatalf("Refresh = %v", err)
	}
	assertFile(t, path, "old")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) { return nil, r.Context().Err() })
	if err := indexer.Refresh(ctx, client, "https://records.test", path); !errors.Is(err, context.Canceled) {
		t.Fatalf("Refresh = %v", err)
	}
	assertFile(t, path, "old")
}

func TestRefreshDoesNotFollowRedirect(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	closed := &trackedBody{Reader: strings.NewReader("redirect")}
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return &http.Response{StatusCode: 302, Header: http.Header{"Location": []string{"https://records.test/next"}}, Body: closed}, nil
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`[]`))}, nil
	})}
	if err := indexer.Refresh(context.Background(), client, "https://records.test", path); err == nil {
		t.Fatal("redirect accepted")
	}
	if calls != 1 || !closed.closed {
		t.Fatalf("calls %d, closed %v", calls, closed.closed)
	}
	assertFile(t, path, "old")
}

func TestRefreshPublicationFailure(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "snapshot")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "keep"), []byte("prior"), 0600); err != nil {
		t.Fatal(err)
	}
	body := &trackedBody{Reader: strings.NewReader(`[]`)}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return &http.Response{StatusCode: 200, Body: body}, nil })}
	if err := indexer.Refresh(context.Background(), client, "https://records.test", path); err == nil {
		t.Fatal("expected publication rejection")
	}
	if !body.closed {
		t.Fatal("body not closed")
	}
	assertFile(t, filepath.Join(path, "keep"), "prior")
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary files remain: %v, %v", entries, err)
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file %s = %q, %v; want %q", path, got, err, want)
	}
}
