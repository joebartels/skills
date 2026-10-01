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

	mirror "example.com/mirror"
)

// A function assignment protects the existing exported signature.
var _ func(context.Context, *http.Client, string, string) error = mirror.Refresh

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type observedBody struct {
	io.Reader
	closes int
}

func (b *observedBody) Close() error {
	b.closes++
	return nil
}

func snapshotPath(t *testing.T, previous string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "items.json")
	if err := os.WriteFile(path, []byte(previous), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertSnapshot(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("snapshot = %q, want %q", got, want)
	}
}

func assertOnlySnapshot(t *testing.T, path string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(path) {
		t.Fatalf("snapshot directory entries = %v, want only %s", entries, filepath.Base(path))
	}
}

func TestRefreshPublishesRemoteSnapshots(t *testing.T) {
	responses := make(chan string, 3)
	responses <- `[{"Code":" A ","Label":" Alpha "},{"code":"B","label":"Beta"}]`
	responses <- `[]`
	responses <- `[{"code":"C","label":"Gamma"}]` + "\n\t "
	requests := make(chan string, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Method + " " + r.URL.RequestURI() + " " + r.Header.Get("X-Snapshot-Client")
		select {
		case body := <-responses:
			_, _ = io.WriteString(w, body)
		default:
			http.Error(w, "unexpected request", http.StatusInternalServerError)
		}
	}))
	defer server.Close()
	transport := server.Client().Transport
	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			r.Header.Set("X-Snapshot-Client", "supplied")
			return transport.RoundTrip(r)
		}),
	}
	path := snapshotPath(t, "previous snapshot\n")
	for _, want := range []string{
		`[{"code":" A ","label":" Alpha "},{"code":"B","label":"Beta"}]`,
		`[]`,
		`[{"code":"C","label":"Gamma"}]`,
	} {
		if err := mirror.Refresh(context.Background(), client, server.URL+"/items?version=2", path); err != nil {
			t.Fatal(err)
		}
		assertSnapshot(t, path, want)
		assertOnlySnapshot(t, path)
		select {
		case request := <-requests:
			if request != "GET /items?version=2 supplied" {
				t.Fatalf("request = %q", request)
			}
		default:
			t.Fatal("supplied HTTP client did not make the request")
		}
	}
}

func TestRefreshRejectsResponsesAndRetainsExactBytes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"non-200", http.StatusServiceUnavailable, `[]`},
		{"201", http.StatusCreated, `[]`},
		{"204", http.StatusNoContent, `[]`},
		{"empty body", http.StatusOK, ``},
		{"malformed JSON", http.StatusOK, `[{"code":"new","label":"New"}`},
		{"trailing JSON", http.StatusOK, `[{"code":"new","label":"New"}] []`},
		{"trailing garbage", http.StatusOK, `[] garbage`},
		{"null is not an array", http.StatusOK, `null`},
		{"object is not an array", http.StatusOK, `{"code":"new","label":"New"}`},
		{"wrong field type", http.StatusOK, `[{"code":12,"label":"New"}]`},
		{"empty code", http.StatusOK, `[{"code":"","label":"New"}]`},
		{"blank code", http.StatusOK, `[{"code":" \t\n\u2003","label":"New"}]`},
		{"empty label", http.StatusOK, `[{"code":"new","label":""}]`},
		{"blank label", http.StatusOK, `[{"code":"new","label":" \t\n\u2003"}]`},
		{"invalid later item", http.StatusOK, `[{"code":"new","label":"New"},{"code":"bad"}]`},
		{"null item", http.StatusOK, `[null]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			previous := "  previous bytes\nwith unusual formatting\t"
			path := snapshotPath(t, previous)
			body := &observedBody{Reader: strings.NewReader(tc.body)}
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: body}, nil
			})}
			if err := mirror.Refresh(context.Background(), client, "http://snapshot.invalid/items", path); err == nil {
				t.Fatal("Refresh accepted an invalid response")
			}
			assertSnapshot(t, path, previous)
			assertOnlySnapshot(t, path)
			if body.closes != 1 {
				t.Fatalf("response body close calls = %d, want 1", body.closes)
			}
		})
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("response stream interrupted")
}

func TestRefreshResponseReadFailure(t *testing.T) {
	previous := "accepted snapshot\n"
	path := snapshotPath(t, previous)
	body := &observedBody{Reader: io.MultiReader(
		strings.NewReader(`[{"code":"new","label":"New"}]`), failingReader{},
	)}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
	})}
	if err := mirror.Refresh(context.Background(), client, "http://snapshot.invalid/items", path); err == nil {
		t.Fatal("Refresh accepted a response whose read failed after complete JSON")
	}
	assertSnapshot(t, path, previous)
	assertOnlySnapshot(t, path)
	if body.closes != 1 {
		t.Fatalf("response body close calls = %d, want 1", body.closes)
	}
}

func TestRefreshClosesSuccessfulResponse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items.json")
	body := &observedBody{Reader: strings.NewReader(`[]`)}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
	})}
	if err := mirror.Refresh(context.Background(), client, "http://snapshot.invalid/items", path); err != nil {
		t.Fatal(err)
	}
	assertSnapshot(t, path, `[]`)
	if body.closes != 1 {
		t.Fatalf("response body close calls = %d, want 1", body.closes)
	}
}

func TestRefreshTransportFailure(t *testing.T) {
	previous := "previous snapshot\n"
	path := snapshotPath(t, previous)
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("remote unavailable")
	})}
	if err := mirror.Refresh(context.Background(), client, "http://snapshot.invalid/items", path); err == nil {
		t.Fatal("Refresh ignored the transport failure")
	}
	if calls != 1 {
		t.Fatalf("transport calls = %d, want 1", calls)
	}
	assertSnapshot(t, path, previous)
	assertOnlySnapshot(t, path)
}

func TestRefreshCancellationReachesHTTP(t *testing.T) {
	started := make(chan struct{})
	canceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-r.Context().Done():
			close(canceled)
		case <-time.After(4 * time.Second):
			http.Error(w, "request did not cancel", http.StatusGatewayTimeout)
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 3 * time.Second
	previous := "previous snapshot\n"
	path := snapshotPath(t, previous)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- mirror.Refresh(ctx, client, server.URL, path) }()
	select {
	case <-started:
	case err := <-result:
		t.Fatalf("Refresh returned before the request started: %v", err)
	case <-time.After(time.Second):
		t.Fatal("HTTP request did not start")
	}
	cancel()
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("Refresh succeeded after cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("Refresh did not return after cancellation")
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("server did not observe cancellation")
	}
	assertSnapshot(t, path, previous)
	assertOnlySnapshot(t, path)
}

func TestRefreshInvalidEndpoint(t *testing.T) {
	previous := "previous snapshot\n"
	path := snapshotPath(t, previous)
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("invalid endpoint reached the transport")
		return nil, nil
	})}
	if err := mirror.Refresh(context.Background(), client, "http://bad\nhost", path); err == nil {
		t.Fatal("Refresh accepted an invalid endpoint")
	}
	assertSnapshot(t, path, previous)
	assertOnlySnapshot(t, path)
}

func TestRefreshPublicationFailureCleansTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "items.json")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(path, "retained")
	if err := os.WriteFile(marker, []byte("old state"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := &observedBody{Reader: strings.NewReader(`[]`)}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
	})}
	if err := mirror.Refresh(context.Background(), client, "http://snapshot.invalid/items", path); err == nil {
		t.Fatal("Refresh replaced a nonempty directory")
	}
	assertSnapshot(t, marker, "old state")
	assertOnlySnapshot(t, path)
	if body.closes != 1 {
		t.Fatalf("response body close calls = %d, want 1", body.closes)
	}
}

func TestRefreshRetainsLastAcceptedSnapshot(t *testing.T) {
	path := snapshotPath(t, "original snapshot")
	responses := []string{
		`[{"code":"accepted","label":"Accepted"}]`,
		`[{"code":"valid","label":"Valid"},{"code":"invalid","label":" "}]`,
		`[{"code":"valid","label":"Valid"}] []`,
		`[{"code":"later","label":"Later"}]`,
	}
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		body := io.NopCloser(strings.NewReader(responses[calls]))
		calls++
		return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
	})}
	for i := range responses {
		err := mirror.Refresh(context.Background(), client, "http://snapshot.invalid/items", path)
		if i == 1 || i == 2 {
			if err == nil {
				t.Fatalf("replacement %d succeeded, want rejection", i)
			}
		} else if err != nil {
			t.Fatal(err)
		}
		want := `[{"code":"accepted","label":"Accepted"}]`
		if i == 3 {
			want = `[{"code":"later","label":"Later"}]`
		}
		assertSnapshot(t, path, want)
		assertOnlySnapshot(t, path)
	}
}
