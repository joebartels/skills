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

type observedBody struct {
	io.Reader
	closed int
}

func (b *observedBody) Close() error { b.closed++; return nil }

func responseClient(t *testing.T, ctx context.Context, status int, body *observedBody) (*http.Client, *int) {
	t.Helper()
	calls := new(int)
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		(*calls)++
		if r.Method != http.MethodGet || r.URL.String() != "https://records.invalid/feed" || r.Context() != ctx {
			t.Errorf("request = %s %s with context %v; want supplied GET/context", r.Method, r.URL, r.Context())
		}
		return &http.Response{StatusCode: status, Body: body}, nil
	})}
	return client, calls
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file = %q, %v; want %q", got, err, want)
	}
}

func TestRefreshPublishesExactArrayAndRetainsOpenSnapshot(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"ordered records", " [ {\"key\":\"alpha\",\"text\":\"  first\\n \"}, {\"key\":\"alpha\",\"text\":\"two\"}, {\"key\":\"beta\",\"text\":\"é\"} ] \n\t", "[{\"key\":\"alpha\",\"text\":\"  first\\n \"},{\"key\":\"alpha\",\"text\":\"two\"},{\"key\":\"beta\",\"text\":\"é\"}]\n"},
		{"empty", "[]\n", "[]\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "records.json")
			old := "old snapshot with exact\nbytes\x00"
			if err := os.WriteFile(path, []byte(old), 0600); err != nil {
				t.Fatal(err)
			}
			reader, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			body := &observedBody{Reader: strings.NewReader(tc.input)}
			client, calls := responseClient(t, ctx, 200, body)
			if err := indexer.Refresh(ctx, client, "https://records.invalid/feed", path); err != nil {
				t.Fatal(err)
			}
			if *calls != 1 || body.closed != 1 {
				t.Fatalf("requests = %d, body closes = %d; want one of each", *calls, body.closed)
			}
			assertFile(t, path, tc.want)
			got, err := io.ReadAll(reader)
			if err != nil || string(got) != old {
				t.Fatalf("open snapshot = %q, %v; want %q", got, err, old)
			}
		})
	}
}

func TestRefreshRejectsAndRetainsPriorBytes(t *testing.T) {
	readFailure := errors.New("body read failed")
	cases := []struct {
		name   string
		status int
		input  io.Reader
		cause  error
	}{
		{"status", 503, strings.NewReader("[]"), nil},
		{"other success status", 201, strings.NewReader("[]"), nil},
		{"empty body", 200, strings.NewReader(""), nil},
		{"object", 200, strings.NewReader(`{"key":"alpha","text":"one"}`), nil},
		{"null", 200, strings.NewReader("null"), nil},
		{"malformed", 200, strings.NewReader("[{"), nil},
		{"second value", 200, strings.NewReader("[] []"), nil},
		{"trailing junk", 200, strings.NewReader("[] !"), nil},
		{"invalid key", 200, strings.NewReader(`[{"key":"alpha","text":"one"},{"key":"../bad","text":"two"}]`), indexer.ErrInvalidRecord},
		{"blank text", 200, strings.NewReader(`[{"key":"alpha","text":" \t\u2003"}]`), indexer.ErrInvalidRecord},
		{"null record", 200, strings.NewReader("[null]"), indexer.ErrInvalidRecord},
		{"read before data", 200, failingReader{readFailure}, readFailure},
		{"read after array", 200, io.MultiReader(strings.NewReader("[]"), failingReader{readFailure}), readFailure},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "records.json")
			old := "\x00 prior bytes \n"
			if err := os.WriteFile(path, []byte(old), 0600); err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			body := &observedBody{Reader: tc.input}
			client, calls := responseClient(t, ctx, tc.status, body)
			err := indexer.Refresh(ctx, client, "https://records.invalid/feed", path)
			if err == nil || (tc.cause != nil && !errors.Is(err, tc.cause)) {
				t.Fatalf("Refresh = %v; want rejection with cause %v", err, tc.cause)
			}
			assertFile(t, path, old)
			if *calls != 1 || body.closed != 1 {
				t.Fatalf("requests = %d, body closes = %d", *calls, body.closed)
			}
		})
	}
}

func TestRefreshFetchFailureAndCallerCancellation(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		t.Run(map[bool]string{false: "fetch failure", true: "cancellation"}[canceled], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			path := filepath.Join(t.TempDir(), "records.json")
			if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
				t.Fatal(err)
			}
			failure := errors.New("transport failed")
			if canceled {
				failure = context.Canceled
			}
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Context() != ctx {
					t.Error("caller context was replaced")
				}
				if canceled {
					cancel()
					<-r.Context().Done()
				}
				return nil, failure
			})}
			if err := indexer.Refresh(ctx, client, "https://records.invalid/feed", path); !errors.Is(err, failure) {
				t.Fatalf("Refresh = %v; want %v", err, failure)
			}
			if calls != 1 {
				t.Fatalf("requests = %d", calls)
			}
			assertFile(t, path, "prior")
		})
	}
}

func TestRefreshPublicationFailureClosesBodyAndCleansTemporary(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "records.json")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(path, "marker")
	if err := os.WriteFile(marker, []byte("prior directory content"), 0600); err != nil {
		t.Fatal(err)
	}
	body := &observedBody{Reader: strings.NewReader("[]")}
	client, _ := responseClient(t, context.Background(), 200, body)
	if err := indexer.Refresh(context.Background(), client, "https://records.invalid/feed", path); err == nil {
		t.Fatal("publishing over a directory succeeded")
	}
	if body.closed != 1 {
		t.Fatalf("body closes = %d", body.closed)
	}
	assertFile(t, marker, "prior directory content")
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 || entries[0].Name() != "records.json" {
		t.Fatalf("root after rejection = %v, %v; want destination only", entries, err)
	}
}
