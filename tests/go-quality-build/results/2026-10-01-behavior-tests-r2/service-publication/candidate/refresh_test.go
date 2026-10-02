package mirror_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/mirror"
)

// Compile the existing public function type from a consumer package.
var _ func(context.Context, *http.Client, string, string) error = mirror.Refresh

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type trackedBody struct {
	io.Reader
	closes int
}

func (b *trackedBody) Close() error {
	b.closes++
	return nil
}

func responseClient(status int, body *trackedBody) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Body: body, Header: make(http.Header)}, nil
	})}
}

func requireFileBytes(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("snapshot = %q, want %q", got, want)
	}
}

func requireOnlySnapshot(t *testing.T, path string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(path) {
		t.Fatalf("snapshot directory = %v, want only %s", entries, filepath.Base(path))
	}
}

func TestRefreshPublishesExactValues(t *testing.T) {
	for _, tc := range []struct {
		name, input, want string
	}{
		{
			name:  "nonblank values retain whitespace and case",
			input: "[{\"Code\":\" A \",\"Label\":\"  Web  \"},{\"code\":\"β\",\"label\":\"Catalog\"}] \n\t",
			want:  `[{"code":" A ","label":"  Web  "},{"code":"β","label":"Catalog"}]`,
		},
		{name: "empty array", input: `[]`, want: `[]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "items.json")
			body := &trackedBody{Reader: strings.NewReader(tc.input)}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := mirror.Refresh(ctx, responseClient(http.StatusOK, body), "https://mirror.invalid/items", path); err != nil {
				t.Fatal(err)
			}
			requireFileBytes(t, path, tc.want)
			requireOnlySnapshot(t, path)
			if body.closes != 1 {
				t.Errorf("body close calls = %d, want 1", body.closes)
			}
		})
	}
}

func TestRefreshRejectsResponseWithoutChangingPriorBytes(t *testing.T) {
	const old = "[ { \"code\": \"prior\", \"label\": \"Snapshot\" } ]\n"
	const valid = `[{"code":"new","label":"Snapshot"}]`
	for _, tc := range []struct {
		name   string
		status int
		input  string
	}{
		{name: "partial content with valid body", status: http.StatusPartialContent, input: valid},
		{name: "created with valid body", status: http.StatusCreated, input: valid},
		{name: "service unavailable", status: http.StatusServiceUnavailable, input: valid},
		{name: "malformed JSON", status: http.StatusOK, input: `[{`},
		{name: "trailing array", status: http.StatusOK, input: valid + `[]`},
		{name: "trailing null", status: http.StatusOK, input: valid + `null`},
		{name: "trailing junk", status: http.StatusOK, input: valid + ` garbage`},
		{name: "null instead of array", status: http.StatusOK, input: `null`},
		{name: "object instead of array", status: http.StatusOK, input: `{"code":"new","label":"Snapshot"}`},
		{name: "blank code", status: http.StatusOK, input: `[{"code":" \t\n","label":"Valid"}]`},
		{name: "unicode blank label", status: http.StatusOK, input: `[{"code":"valid","label":"\u2003\u00a0"}]`},
		{name: "missing code", status: http.StatusOK, input: `[{"label":"Valid"}]`},
		{name: "missing label", status: http.StatusOK, input: `[{"code":"valid"}]`},
		{name: "wrong field type", status: http.StatusOK, input: `[{"code":4,"label":"Valid"}]`},
		{name: "invalid item after valid prefix", status: http.StatusOK, input: `[{"code":"first","label":"Valid"},{"code":"second","label":""}]`},
		{name: "null item", status: http.StatusOK, input: `[null]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "items.json")
			if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
				t.Fatal(err)
			}
			body := &trackedBody{Reader: strings.NewReader(tc.input)}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := mirror.Refresh(ctx, responseClient(tc.status, body), "https://mirror.invalid/items", path); err == nil {
				t.Fatal("Refresh succeeded for rejected response")
			}
			requireFileBytes(t, path, old)
			requireOnlySnapshot(t, path)
			if body.closes != 1 {
				t.Errorf("body close calls = %d, want 1", body.closes)
			}
		})
	}
}

func TestRefreshUsesSuppliedClientAndContext(t *testing.T) {
	type contextKey struct{}
	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), contextKey{}, "request marker"), 3*time.Second)
	defer cancel()
	body := &trackedBody{Reader: strings.NewReader(`[]`)}
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if req.Method != http.MethodGet || req.URL.String() != "https://mirror.invalid/items?batch=2" {
			t.Errorf("request = %s %s, want GET supplied endpoint", req.Method, req.URL)
		}
		if req.Context().Value(contextKey{}) != "request marker" {
			t.Error("request lost supplied context value")
		}
		gotDeadline, gotOK := req.Context().Deadline()
		wantDeadline, _ := ctx.Deadline()
		if !gotOK || !gotDeadline.Equal(wantDeadline) {
			t.Errorf("request deadline = %v, %t, want %v", gotDeadline, gotOK, wantDeadline)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
	})}
	path := filepath.Join(t.TempDir(), "items.json")
	if err := mirror.Refresh(ctx, client, "https://mirror.invalid/items?batch=2", path); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("supplied transport calls = %d, want 1", calls)
	}
	requireFileBytes(t, path, `[]`)
	if body.closes != 1 {
		t.Errorf("body close calls = %d, want 1", body.closes)
	}
}

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }

func TestRefreshRequestAndReadFailuresRetainSnapshot(t *testing.T) {
	for _, failure := range []string{"transport", "invalid endpoint", "canceled context", "body read", "read after JSON"} {
		t.Run(failure, func(t *testing.T) {
			const old = "previous exact bytes\n"
			path := filepath.Join(t.TempDir(), "items.json")
			if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			endpoint := "https://mirror.invalid/items"
			body := &trackedBody{Reader: strings.NewReader(`[]`)}
			transportErr := errors.New("transport failure")
			readErr := errors.New("response read failure")
			if failure == "invalid endpoint" {
				endpoint = "://invalid"
			}
			if failure == "canceled context" {
				cancel()
			}
			if failure == "body read" {
				body.Reader = failedReader{err: readErr}
			}
			if failure == "read after JSON" {
				body.Reader = io.MultiReader(strings.NewReader(`[]`), failedReader{err: readErr})
			}
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if failure == "transport" {
					return nil, transportErr
				}
				if failure == "invalid endpoint" {
					t.Error("invalid endpoint reached transport")
				}
				if failure == "canceled context" {
					select {
					case <-req.Context().Done():
						return nil, req.Context().Err()
					case <-time.After(time.Second):
						t.Error("supplied cancellation was not forwarded")
						return nil, transportErr
					}
				}
				return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
			})}
			if err := mirror.Refresh(ctx, client, endpoint, path); err == nil {
				t.Fatal("Refresh succeeded after failure")
			}
			requireFileBytes(t, path, old)
			requireOnlySnapshot(t, path)
			wantCloses := 0
			if failure == "body read" || failure == "read after JSON" {
				wantCloses = 1
			}
			if body.closes != wantCloses {
				t.Errorf("body close calls = %d, want %d", body.closes, wantCloses)
			}
		})
	}
}

func TestRefreshThroughHTTP(t *testing.T) {
	methods := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		methods <- req.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `[{"code":"remote","label":"Remote snapshot"}]`)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	path := filepath.Join(t.TempDir(), "items.json")
	if err := mirror.Refresh(ctx, server.Client(), server.URL, path); err != nil {
		t.Fatal(err)
	}
	select {
	case method := <-methods:
		if method != http.MethodGet {
			t.Errorf("HTTP method = %s, want GET", method)
		}
	case <-ctx.Done():
		t.Fatal("server did not observe request")
	}
	requireFileBytes(t, path, `[{"code":"remote","label":"Remote snapshot"}]`)
}

func TestRefreshRenameFailureClosesBodyAndCleansStage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "items.json")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(path, "keep")
	if err := os.WriteFile(marker, []byte("retained"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := &trackedBody{Reader: strings.NewReader(`[]`)}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := mirror.Refresh(ctx, responseClient(http.StatusOK, body), "https://mirror.invalid/items", path); err == nil {
		t.Fatal("Refresh succeeded replacing nonempty directory")
	}
	requireFileBytes(t, marker, "retained")
	requireOnlySnapshot(t, path)
	if body.closes != 1 {
		t.Errorf("body close calls = %d, want 1", body.closes)
	}
}
