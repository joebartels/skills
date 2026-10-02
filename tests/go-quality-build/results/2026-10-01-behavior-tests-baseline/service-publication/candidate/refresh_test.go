package mirror_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"example.com/mirror"
)

// The documented function type remains available to external callers.
var _ func(context.Context, *http.Client, string, string) error = mirror.Refresh

func TestRefreshPublishesRemoteSnapshot(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want string
	}{
		{
			name: "values and order preserved",
			body: `[{"code":"  x\t","label":" Label \n"},{"code":"x","label":"Other"}]` + " \n\t",
			want: `[{"code":"  x\t","label":" Label \n"},{"code":"x","label":"Other"}]`,
		},
		{name: "empty array", body: `[]`, want: `[]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := make(chan string, 1)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- r.Method + " " + r.URL.RequestURI()
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, tc.body)
			})
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, r)
				return response.Result(), nil
			})}
			client.Timeout = 2 * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			path := filepath.Join(t.TempDir(), "new.json")
			if err := mirror.Refresh(ctx, client, "https://example.invalid/items?view=all", path); err != nil {
				t.Fatalf("Refresh: %v", err)
			}
			assertSnapshot(t, path, []byte(tc.want))
			select {
			case got := <-requests:
				if got != "GET /items?view=all" {
					t.Fatalf("request = %q", got)
				}
			default:
				t.Fatal("supplied client's handler received no request")
			}
		})
	}
}

func TestRefreshFailuresPreserveSnapshotAndCloseBody(t *testing.T) {
	readFailure := errors.New("response read failed")
	for _, tc := range []struct {
		name   string
		status int
		body   string
		reader io.Reader
		cause  error
	}{
		{name: "created status", status: http.StatusCreated, body: `[{"code":"x","label":"X"}]`},
		{name: "no content status", status: http.StatusNoContent},
		{name: "not found status", status: http.StatusNotFound, body: `[]`},
		{name: "server error status", status: http.StatusInternalServerError, body: `[]`},
		{name: "empty response"},
		{name: "malformed JSON", body: `[{"code":`},
		{name: "wrong item field type", body: `[{"code":42,"label":"X"}]`},
		{name: "null document", body: `null`},
		{name: "object document", body: `{"code":"x","label":"X"}`},
		{name: "null item", body: `[null]`},
		{name: "blank code", body: `[{"code":" \t\n","label":"X"}]`},
		{name: "blank label", body: `[{"code":"x","label":"\u2003\r "}]`},
		{name: "missing code", body: `[{"label":"X"}]`},
		{name: "missing label", body: `[{"code":"x"}]`},
		{name: "later invalid item", body: `[{"code":"x","label":"X"},{"code":"","label":"Y"}]`},
		{name: "second array", body: `[] []`},
		{name: "trailing scalar", body: `[] true`},
		{name: "trailing malformed JSON", body: `[] garbage`},
		{name: "initial read failure", reader: errorReader{readFailure}, cause: readFailure},
		{name: "read failure after array", reader: io.MultiReader(strings.NewReader(`[]`), errorReader{readFailure}), cause: readFailure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path, prior := existingSnapshot(t)
			reader := tc.reader
			if reader == nil {
				reader = strings.NewReader(tc.body)
			}
			body := &trackedBody{Reader: reader}
			status := tc.status
			if status == 0 {
				status = http.StatusOK
			}
			client := responseClient(status, body)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			err := mirror.Refresh(ctx, client, "https://example.invalid/items", path)
			if err == nil {
				t.Fatal("Refresh unexpectedly succeeded")
			}
			if tc.cause != nil && !errors.Is(err, tc.cause) {
				t.Fatalf("Refresh error = %v, want cause %v", err, tc.cause)
			}
			assertSnapshot(t, path, prior)
			if body.closes != 1 {
				t.Fatalf("response body close count = %d, want 1", body.closes)
			}
			assertOnlyEntry(t, filepath.Dir(path), filepath.Base(path))
		})
	}
}

func TestRefreshClosesSuccessfulResponseBody(t *testing.T) {
	path, _ := existingSnapshot(t)
	body := &trackedBody{Reader: strings.NewReader(`[{"code":"x","label":"X"}]`)}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := mirror.Refresh(ctx, responseClient(http.StatusOK, body), "https://example.invalid/items", path); err != nil {
		t.Fatal(err)
	}
	assertSnapshot(t, path, []byte(`[{"code":"x","label":"X"}]`))
	if body.closes != 1 {
		t.Fatalf("response body close count = %d, want 1", body.closes)
	}
	assertOnlyEntry(t, filepath.Dir(path), filepath.Base(path))
}

func TestRefreshTransportFailurePreservesSnapshot(t *testing.T) {
	path, prior := existingSnapshot(t)
	wantErr := errors.New("transport unavailable")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, wantErr
	})}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := mirror.Refresh(ctx, client, "https://example.invalid/items", path)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Refresh error = %v, want cause %v", err, wantErr)
	}
	assertSnapshot(t, path, prior)
}

func TestRefreshInvalidRequestPreservesSnapshot(t *testing.T) {
	path, prior := existingSnapshot(t)
	called := false
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		called = true
		return nil, errors.New("unexpected transport call")
	})}
	if err := mirror.Refresh(context.Background(), client, ":invalid", path); err == nil {
		t.Fatal("Refresh accepted an invalid endpoint")
	}
	if called {
		t.Fatal("invalid request reached transport")
	}
	assertSnapshot(t, path, prior)
}

func TestRefreshPropagatesContextCancellation(t *testing.T) {
	path, prior := existingSnapshot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		cancel()
		select {
		case <-r.Context().Done():
			return nil, r.Context().Err()
		case <-time.After(time.Second):
			return nil, errors.New("request did not carry the supplied context")
		}
	})}
	err := mirror.Refresh(ctx, client, "https://example.invalid/items", path)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Refresh error = %v, want context.Canceled", err)
	}
	assertSnapshot(t, path, prior)
}

func TestRefreshReplacesSnapshotForOpenReaders(t *testing.T) {
	switch runtime.GOOS {
	case "aix", "android", "darwin", "dragonfly", "freebsd", "illumos", "ios", "linux", "netbsd", "openbsd", "solaris":
	default:
		t.Skip("open-handle replacement contract applies to supported Unix filesystems")
	}
	path, prior := existingSnapshot(t)
	old, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = old.Close() })
	oldInfo, err := old.Stat()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := mirror.Refresh(ctx, responseClient(http.StatusOK, io.NopCloser(strings.NewReader(`[]`))), "https://example.invalid/items", path); err != nil {
		t.Fatal(err)
	}
	gotOld, err := io.ReadAll(old)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotOld, prior) {
		t.Fatalf("already-open handle = %q, want %q", gotOld, prior)
	}
	assertSnapshot(t, path, []byte(`[]`))
	newInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(oldInfo, newInfo) {
		t.Fatal("Refresh retained the old inode instead of replacing it")
	}
}

func TestRefreshPublicationFailureCleansTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snapshot")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(path, "previous.json")
	prior := []byte("previous snapshot\n")
	if err := os.WriteFile(marker, prior, 0o644); err != nil {
		t.Fatal(err)
	}
	body := &trackedBody{Reader: strings.NewReader(`[]`)}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := mirror.Refresh(ctx, responseClient(http.StatusOK, body), "https://example.invalid/items", path); err == nil {
		t.Fatal("Refresh replaced a nonempty directory")
	}
	assertSnapshot(t, marker, prior)
	assertOnlyEntry(t, dir, "snapshot")
	if body.closes != 1 {
		t.Fatalf("response body close count = %d, want 1", body.closes)
	}
}

func TestRefreshMissingParentReturnsError(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "missing", "items.json")
	body := &trackedBody{Reader: strings.NewReader(`[]`)}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := mirror.Refresh(ctx, responseClient(http.StatusOK, body), "https://example.invalid/items", path); err == nil {
		t.Fatal("Refresh succeeded without a parent directory")
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("directory entries = %v, error = %v", entries, err)
	}
	if body.closes != 1 {
		t.Fatalf("response body close count = %d, want 1", body.closes)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func responseClient(status int, body io.ReadCloser) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Body: body, Header: make(http.Header)}, nil
	})}
}

type trackedBody struct {
	io.Reader
	closes int
}

func (b *trackedBody) Close() error {
	b.closes++
	return nil
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

func existingSnapshot(t *testing.T) (string, []byte) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "items.json")
	prior := []byte("  [ {\"code\":\"old\",\"label\":\"Old\"} ]\n")
	if err := os.WriteFile(path, prior, 0o644); err != nil {
		t.Fatal(err)
	}
	return path, prior
}

func assertSnapshot(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("snapshot = %q, want %q", got, want)
	}
}

func assertOnlyEntry(t *testing.T, dir, want string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != want {
		t.Fatalf("directory entries = %v, want only %q", entries, want)
	}
}
