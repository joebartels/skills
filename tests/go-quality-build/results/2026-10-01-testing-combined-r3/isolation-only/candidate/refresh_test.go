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

type trackedBody struct {
	io.Reader
	closes int
}

func (b *trackedBody) Close() error { b.closes++; return nil }

func TestRefreshValidationAndBodyOwnership(t *testing.T) {
	t.Parallel()
	readErr := errors.New("body read failed")
	cases := []struct {
		name, payload, want string
		status              int
		reader              io.Reader
		invalid             bool
		cause               error
	}{
		{name: "records", status: 200, payload: `[ {"key":"beta","text":"  original \n"}, {"key":"alpha","text":"last"} ] `, want: "[{\"key\":\"beta\",\"text\":\"  original \\n\"},{\"key\":\"alpha\",\"text\":\"last\"}]\n"},
		{name: "empty_array", status: 200, payload: "[]\n\t", want: "[]\n"},
		{name: "status", status: 503, payload: `[]`},
		{name: "created_status", status: 201, payload: `[]`},
		{name: "null", status: 200, payload: `null`},
		{name: "object", status: 200, payload: `{}`},
		{name: "syntax", status: 200, payload: `[{`},
		{name: "second_array", status: 200, payload: `[] []`},
		{name: "trailing_garbage", status: 200, payload: `[] garbage`},
		{name: "invalid_key", status: 200, payload: `[{"key":"Alpha","text":"ok"}]`, invalid: true},
		{name: "blank_text", status: 200, payload: `[{"key":"alpha","text":" \t\n"}]`, invalid: true},
		{name: "late_invalid", status: 200, payload: `[{"key":"alpha","text":"ok"},{"key":"beta","text":""}]`, invalid: true},
		{name: "read_failure", status: 200, reader: failingReader{readErr}, cause: readErr},
		{name: "read_failure_after_array", status: 200, reader: io.MultiReader(strings.NewReader(`[]`), failingReader{readErr}), cause: readErr},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			path := filepath.Join(root, "snapshot")
			prior := "exact prior bytes\x00\n"
			if err := os.WriteFile(path, []byte(prior), 0600); err != nil {
				t.Fatal(err)
			}
			reader := tc.reader
			if reader == nil {
				reader = strings.NewReader(tc.payload)
			}
			body := &trackedBody{Reader: reader}
			ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), struct{}{}, "marker"), 5*time.Second)
			defer cancel()
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				calls++
				if request.Method != http.MethodGet || request.URL.String() != "http://example.test/snapshot" || request.Context() != ctx {
					t.Errorf("request = %s %s, context forwarded %v", request.Method, request.URL, request.Context() == ctx)
				}
				return &http.Response{StatusCode: tc.status, Status: http.StatusText(tc.status), Body: body, Header: make(http.Header)}, nil
			})}
			err := indexer.Refresh(ctx, client, "http://example.test/snapshot", path)
			if (err == nil) != (tc.want != "") {
				t.Fatalf("Refresh error = %v; want success %v", err, tc.want != "")
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("validation identity = %v", err)
			}
			if tc.cause != nil && !errors.Is(err, tc.cause) {
				t.Fatalf("read identity = %v", err)
			}
			if calls != 1 || body.closes != 1 {
				t.Fatalf("requests = %d, body closes = %d", calls, body.closes)
			}
			want := tc.want
			if want == "" {
				want = prior
			}
			got, readError := os.ReadFile(path)
			if readError != nil || string(got) != want {
				t.Fatalf("snapshot = %q, %v; want %q", got, readError, want)
			}
			assertNoTemps(t, root)
		})
	}
}

func TestRefreshFetchAndPublicationFailures(t *testing.T) {
	t.Parallel()
	fetchErr := errors.New("fetch interrupted")
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, fetchErr })}
	if err := indexer.Refresh(context.Background(), client, "http://example.test/", path); !errors.Is(err, fetchErr) {
		t.Fatalf("fetch error = %v", err)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "old" {
		t.Fatalf("prior file = %q, %v", got, err)
	}
	root := t.TempDir()
	blocked := filepath.Join(root, "snapshot")
	if err := os.Mkdir(blocked, 0700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(blocked, "old")
	if err := os.WriteFile(marker, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	body := &trackedBody{Reader: strings.NewReader(`[]`)}
	client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) { return &http.Response{StatusCode: 200, Body: body}, nil })
	if err := indexer.Refresh(context.Background(), client, "http://example.test/", blocked); err == nil {
		t.Fatal("expected rename failure")
	}
	if got, err := os.ReadFile(marker); err != nil || string(got) != "old" {
		t.Fatalf("prior marker = %q, %v", got, err)
	}
	if body.closes != 1 {
		t.Fatalf("body closes = %d", body.closes)
	}
	assertNoTemps(t, root)
}

func TestRefreshPOSIXSnapshot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX open-reader replacement contract")
	}
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "snapshot")
	old := strings.Repeat("old snapshot\n", 1000)
	if err := os.WriteFile(path, []byte(old), 0600); err != nil {
		t.Fatal(err)
	}
	reader, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := reader.Close(); err != nil {
			t.Error(err)
		}
	})
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`[{"key":"alpha","text":"new"}]`))}, nil
	})}
	if err := indexer.Refresh(context.Background(), client, "http://example.test/", path); err != nil {
		t.Fatal(err)
	}
	gotOld, err := io.ReadAll(reader)
	if err != nil || string(gotOld) != old {
		t.Fatalf("open reader lost old snapshot: %v (%d bytes)", err, len(gotOld))
	}
	gotNew, err := os.ReadFile(path)
	if err != nil || string(gotNew) != "[{\"key\":\"alpha\",\"text\":\"new\"}]\n" {
		t.Fatalf("later snapshot = %q, %v", gotNew, err)
	}
}

func TestRefreshRejectsRedirectWithoutChangingClient(t *testing.T) {
	t.Parallel()
	var redirected atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/target", http.StatusFound)
			return
		}
		redirected.Add(1)
		io.WriteString(w, `[]`)
	}))
	t.Cleanup(server.Close)
	var policyCalls atomic.Int32
	client := server.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { policyCalls.Add(1); return nil }
	t.Cleanup(client.CloseIdleConnections)
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := indexer.Refresh(ctx, client, server.URL+"/start", path); err == nil {
		t.Fatal("accepted redirect")
	}
	if redirected.Load() != 0 || policyCalls.Load() != 0 {
		t.Fatalf("followed redirect: target=%d policy=%d", redirected.Load(), policyCalls.Load())
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "old" {
		t.Fatalf("prior = %q, %v", got, err)
	}
	response, err := client.Get(server.URL + "/start")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if redirected.Load() != 1 || policyCalls.Load() != 1 {
		t.Fatal("caller redirect policy changed")
	}
}

func TestRefreshRealRequestCancellation(t *testing.T) {
	t.Parallel()
	started := make(chan struct{})
	stopped := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		close(stopped)
	}))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	client := server.Client()
	client.Timeout = 5 * time.Second
	result := make(chan error, 1)
	joined := make(chan struct{})
	t.Cleanup(func() {
		cancel()
		select {
		case <-joined:
		case <-time.After(6 * time.Second):
			t.Error("refresh did not join")
		}
		client.CloseIdleConnections()
		server.Close()
	})
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	go func() { defer close(joined); result <- indexer.Refresh(ctx, client, server.URL, path) }()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation error = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("request did not stop")
	}
	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		t.Fatal("server did not see cancellation")
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "old" {
		t.Fatalf("prior = %q, %v", got, err)
	}
}
