package indexer_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/indexer"
)

var _ func(context.Context, *http.Client, string, string) error = indexer.Refresh

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackedBody struct {
	io.Reader
	closes int
}

func (b *trackedBody) Close() error { b.closes++; return nil }

func bodyClient(body *trackedBody, status int) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Body: body, Header: make(http.Header)}, nil
	})}
}

func TestRecordJSONContract(t *testing.T) {
	data, err := json.Marshal(indexer.Record{Key: "alpha", Text: "one"})
	if err != nil || string(data) != `{"key":"alpha","text":"one"}` {
		t.Fatalf("JSON = %s, %v", data, err)
	}
}

func TestRefreshPublishesSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "records.json")
	old := "old snapshot with exact whitespace\n\n"
	if err := os.WriteFile(path, []byte(old), 0600); err != nil {
		t.Fatal(err)
	}
	opened, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	body := &trackedBody{Reader: strings.NewReader(` [ {"key":"beta","text":"  spaced  "}, {"key":"alpha","text":"one"}, {"key":"beta","text":"last"} ] `)}
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "marker")
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls++
		if request.Method != "GET" || request.URL.String() != "https://example.test/records" || request.Context() != ctx {
			t.Errorf("request = %s %s, context forwarded = %v", request.Method, request.URL, request.Context() == ctx)
		}
		return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}, nil
	})}
	if err := indexer.Refresh(ctx, client, "https://example.test/records", path); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || body.closes != 1 {
		t.Fatalf("requests = %d, closes = %d", calls, body.closes)
	}
	assertFile(t, path, "[{\"key\":\"beta\",\"text\":\"  spaced  \"},{\"key\":\"alpha\",\"text\":\"one\"},{\"key\":\"beta\",\"text\":\"last\"}]\n")
	got, err := io.ReadAll(opened)
	if err != nil || string(got) != old {
		t.Fatalf("opened snapshot = %q, %v", got, err)
	}
	empty := &trackedBody{Reader: strings.NewReader("[]\n \t")}
	if err := indexer.Refresh(ctx, bodyClient(empty, 200), "https://example.test/records", path); err != nil {
		t.Fatal(err)
	}
	assertFile(t, path, "[]\n")
	if empty.closes != 1 {
		t.Fatalf("empty body closes = %d", empty.closes)
	}
}

func TestRefreshRejectsAndRetainsDestination(t *testing.T) {
	for _, tc := range []struct {
		name, data string
		status     int
		invalid    bool
	}{
		{"status", `[]`, 201, false},
		{"server error", `[]`, 503, false},
		{"null", `null`, 200, false},
		{"object", `{}`, 200, false},
		{"malformed", `[{`, 200, false},
		{"second value", `[] []`, 200, false},
		{"trailing garbage", `[] x`, 200, false},
		{"invalid key", `[{"key":"Alpha","text":"one"}]`, 200, true},
		{"blank text", `[{"key":"alpha","text":" \t\n"}]`, 200, true},
		{"bad later record", `[{"key":"alpha","text":"one"},{"key":"beta","text":""}]`, 200, true},
		{"wrong text type", `[{"key":"alpha","text":1}]`, 200, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "records.json")
			old := "prior bytes \n"
			if err := os.WriteFile(path, []byte(old), 0600); err != nil {
				t.Fatal(err)
			}
			body := &trackedBody{Reader: strings.NewReader(tc.data)}
			err := indexer.Refresh(context.Background(), bodyClient(body, tc.status), "https://example.test/records", path)
			if err == nil {
				t.Fatal("missing rejection")
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("error = %v, want invalid record", err)
			}
			assertFile(t, path, old)
			if body.closes != 1 {
				t.Fatalf("body closes = %d", body.closes)
			}
			entries, err := os.ReadDir(filepath.Dir(path))
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 {
				t.Fatalf("leaked temporary files: %v", entries)
			}
		})
	}
}

func TestRefreshIOFailures(t *testing.T) {
	cause := errors.New("fetch/read interrupted")
	for _, stage := range []string{"fetch", "read", "read after JSON", "publish"} {
		t.Run(stage, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "records.json")
			old := "prior\n"
			if err := os.WriteFile(path, []byte(old), 0600); err != nil {
				t.Fatal(err)
			}
			body := &trackedBody{Reader: strings.NewReader(`[{"key":"alpha","text":"one"}]`)}
			client := bodyClient(body, 200)
			switch stage {
			case "fetch":
				client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, cause })
			case "read":
				body.Reader = failingReader{cause}
			case "read after JSON":
				body.Reader = io.MultiReader(body.Reader, failingReader{cause})
			case "publish":
				path = filepath.Dir(path)
			}
			err := indexer.Refresh(context.Background(), client, "https://example.test/records", path)
			if err == nil {
				t.Fatal("missing rejection")
			}
			if stage != "publish" && !errors.Is(err, cause) {
				t.Fatalf("error = %v, want cause", err)
			}
			if stage == "publish" {
				assertFile(t, filepath.Join(path, "records.json"), old)
			} else {
				assertFile(t, path, old)
			}
			wantCloses := 1
			if stage == "fetch" {
				wantCloses = 0
			}
			if body.closes != wantCloses {
				t.Fatalf("closes = %d, want %d", body.closes, wantCloses)
			}
		})
	}
}

func TestRefreshContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "records.json")
	if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Context() != ctx {
			t.Error("context was replaced")
		}
		cancel()
		<-request.Context().Done()
		return nil, request.Context().Err()
	})}
	if err := indexer.Refresh(ctx, client, "https://example.test/records", path); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
	assertFile(t, path, "prior")
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file = %q, %v; want %q", got, err, want)
	}
}
