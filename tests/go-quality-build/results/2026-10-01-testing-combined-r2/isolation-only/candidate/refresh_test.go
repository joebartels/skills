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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type observedBody struct {
	io.Reader
	closes int
}

func (b *observedBody) Close() error { b.closes++; return nil }

func TestRefreshResponseContracts(t *testing.T) {
	readErr := errors.New("body read failed")
	cases := []struct {
		name    string
		status  int
		body    string
		tailErr error
		want    string
		invalid bool
	}{
		{"ordered_text_and_duplicates", 200, `[{"key":"beta","text":" original \n"},{"key":"alpha","text":"one"},{"key":"alpha","text":"two"}]`, nil, "[{\"key\":\"beta\",\"text\":\" original \\n\"},{\"key\":\"alpha\",\"text\":\"one\"},{\"key\":\"alpha\",\"text\":\"two\"}]\n", false},
		{"empty", 200, "[] \n\t", nil, "[]\n", false},
		{"null", 200, "null", nil, "", false},
		{"object", 200, `{ "key": "alpha", "text": "one" }`, nil, "", false},
		{"truncated", 200, `[{"key":"alpha"`, nil, "", false},
		{"trailing_value", 200, "[] []", nil, "", false},
		{"trailing_junk", 200, "[] junk", nil, "", false},
		{"invalid_key", 200, `[{"key":"Alpha","text":"one"}]`, nil, "", true},
		{"blank_text", 200, `[{"key":"alpha","text":" \t\n\u2003"}]`, nil, "", true},
		{"missing_text", 200, `[{"key":"alpha"}]`, nil, "", true},
		{"late_validation", 200, `[{"key":"alpha","text":"good"},{"key":"beta","text":""}]`, nil, "", true},
		{"read_before_data", 200, "", readErr, "", false},
		{"read_after_array", 200, "[]", readErr, "", false},
		{"redirect", 302, "[]", nil, "", false},
		{"other_success_status", 204, "[]", nil, "", false},
		{"server_error", 500, "[]", nil, "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "snapshot.json")
			prior := "exact prior bytes\x00\n"
			if err := os.WriteFile(path, []byte(prior), 0600); err != nil {
				t.Fatal(err)
			}
			var reader io.Reader = strings.NewReader(tc.body)
			if tc.tailErr != nil {
				reader = io.MultiReader(reader, errorReader{tc.tailErr})
			}
			body := &observedBody{Reader: reader}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			requests := 0
			var requestErr error
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				requests++
				if r.Method != http.MethodGet || r.URL.String() != "https://source.example/snapshot" || r.Context() != ctx {
					requestErr = errors.New("request method, endpoint or context changed")
				}
				return &http.Response{StatusCode: tc.status, Body: body, Header: make(http.Header), Request: r}, nil
			})}
			err := indexer.Refresh(ctx, client, "https://source.example/snapshot", path)
			if requests != 1 || body.closes != 1 || requestErr != nil {
				t.Fatalf("requests=%d, closes=%d, request error=%v", requests, body.closes, requestErr)
			}
			if tc.want != "" {
				if err != nil {
					t.Fatal(err)
				}
				assertFile(t, path, tc.want)
				return
			}
			if err == nil || errors.Is(err, indexer.ErrInvalidRecord) != tc.invalid {
				t.Fatalf("Refresh = %v; validation identity want %v", err, tc.invalid)
			}
			if tc.tailErr != nil && !errors.Is(err, tc.tailErr) {
				t.Fatalf("read error identity lost: %v", err)
			}
			assertFile(t, path, prior)
		})
	}
}

func TestRefreshFetchFailureRetainsDestination(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("transport failed")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, wantErr })}
	if err := indexer.Refresh(context.Background(), client, "http://source.example", path); !errors.Is(err, wantErr) {
		t.Fatalf("Refresh = %v; want transport error", err)
	}
	assertFile(t, path, "prior")
}

func TestRefreshRedirectDoesNotFollowOrMutateClient(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		io.WriteString(w, "[]")
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	var policyCalls atomic.Int32
	policyErr := errors.New("caller redirect policy")
	client.CheckRedirect = func(*http.Request, []*http.Request) error { policyCalls.Add(1); return policyErr }
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := indexer.Refresh(ctx, client, server.URL+"/redirect", path); err == nil {
		t.Fatal("redirect accepted")
	}
	if requests.Load() != 1 || policyCalls.Load() != 0 {
		t.Fatalf("requests=%d, caller policy calls=%d", requests.Load(), policyCalls.Load())
	}
	if err := client.CheckRedirect(nil, nil); !errors.Is(err, policyErr) {
		t.Fatalf("caller policy changed: %v", err)
	}
	assertFile(t, path, "prior")
}

func TestRefreshNetworkCancellation(t *testing.T) {
	t.Parallel()
	started, stopped := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		close(stopped)
	}))
	t.Cleanup(server.Close)
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	done := make(chan error, 1)
	finished := make(chan struct{})
	t.Cleanup(func() { cancel(); await(t, finished, "refresh cancellation cleanup") })
	go func() {
		defer close(finished)
		done <- indexer.Refresh(ctx, server.Client(), server.URL, path)
	}()
	await(t, started, "HTTP request startup")
	cancel()
	if err := await(t, done, "refresh cancellation"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Refresh = %v; want canceled", err)
	}
	await(t, stopped, "server request cancellation")
	assertFile(t, path, "prior")
}

func TestRefreshAtomicSnapshot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX replacement contract")
	}
	t.Parallel()
	path := filepath.Join(t.TempDir(), "snapshot")
	old := strings.Repeat("old snapshot\n", 1000)
	if err := os.WriteFile(path, []byte(old), 0600); err != nil {
		t.Fatal(err)
	}
	reader, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { reader.Close() })
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`[{"key":"alpha","text":"new"}]`)), Request: r}, nil
	})}
	if err := indexer.Refresh(context.Background(), client, "http://source.example", path); err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(reader)
	if err != nil || string(got) != old {
		t.Fatalf("old reader = %q, %v", got, err)
	}
	assertFile(t, path, "[{\"key\":\"alpha\",\"text\":\"new\"}]\n")
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 1 {
		t.Fatalf("publication files = %v, %v", entries, err)
	}
}

func TestRefreshPublicationFailureClosesBody(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "blocked")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(path, "prior")
	if err := os.WriteFile(marker, []byte("prior"), 0600); err != nil {
		t.Fatal(err)
	}
	body := &observedBody{Reader: strings.NewReader("[]")}
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body, Request: r}, nil
	})}
	if err := indexer.Refresh(context.Background(), client, "http://source.example", path); err == nil {
		t.Fatal("publication succeeded over directory")
	}
	if body.closes != 1 {
		t.Fatalf("closes = %d", body.closes)
	}
	assertFile(t, marker, "prior")
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 1 {
		t.Fatalf("publication files = %v, %v", entries, err)
	}
}

func TestRefreshWriteFailurePreservesPriorBytes(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "snapshot")
	if err := os.WriteFile(path, []byte("prior\x00\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(root, 0700) })
	body := &observedBody{Reader: strings.NewReader("[]")}
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: body, Request: r}, nil
	})}
	err := indexer.Refresh(context.Background(), client, "http://source.example", path)
	if err == nil {
		t.Skip("filesystem privileges bypass directory write permissions")
	}
	if body.closes != 1 {
		t.Fatalf("closes = %d", body.closes)
	}
	assertFile(t, path, "prior\x00\n")
}
