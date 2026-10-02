package indexer_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
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
	reader io.Reader
	closes int
}

func (b *observedBody) Read(p []byte) (int, error) { return b.reader.Read(p) }
func (b *observedBody) Close() error               { b.closes++; return nil }

type brokenReader struct{ err error }

func (r brokenReader) Read(p []byte) (int, error) {
	copy(p, "[")
	return 1, r.err
}

func TestRefreshPublishesKnownBytes(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"ordered original text", ` [ {"text":"  first  ","key":"beta"}, {"key":"alpha","text":"second"}, {"key":"beta","text":"last"} ] `, "[{\"key\":\"beta\",\"text\":\"  first  \"},{\"key\":\"alpha\",\"text\":\"second\"},{\"key\":\"beta\",\"text\":\"last\"}]\n"},
		{"empty array", "[]\n\t", "[]\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			body := &observedBody{reader: strings.NewReader(tc.input)}
			requests := 0
			ctx := context.WithValue(context.Background(), struct{}{}, "caller")
			client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				requests++
				if request.Method != http.MethodGet || request.URL.String() != "https://example.test/source" || request.Context() != ctx {
					return nil, errors.New("request method, URL or caller context changed")
				}
				return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header)}, nil
			})}
			path := filepath.Join(t.TempDir(), "snapshot")
			if err := os.WriteFile(path, []byte("old snapshot\x00\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := indexer.Refresh(ctx, client, "https://example.test/source", path); err != nil {
				t.Fatal(err)
			}
			assertFile(t, path, tc.want)
			if requests != 1 || body.closes != 1 {
				t.Fatalf("requests/closes = %d/%d; want 1/1", requests, body.closes)
			}
			assertNoTemps(t, filepath.Dir(path))
		})
	}
	if fields := reflect.TypeOf(indexer.Record{}); fields.Field(0).Tag.Get("json") != "key" || fields.Field(1).Tag.Get("json") != "text" {
		t.Fatal("Record JSON field contract changed")
	}
}

func TestRefreshRejectionsKeepExactPriorBytes(t *testing.T) {
	readErr, fetchErr := errors.New("partial body read failed"), errors.New("transport failed")
	for _, tc := range []struct {
		name              string
		status            int
		input             string
		readErr, fetchErr error
		invalid           bool
	}{
		{name: "status", status: 503, input: `[]`},
		{name: "invalid key", status: 200, input: `[{"key":"ALPHA","text":"new"}]`, invalid: true},
		{name: "blank text", status: 200, input: `[{"key":"alpha","text":" \t"}]`, invalid: true},
		{name: "late invalid record", status: 200, input: `[{"key":"alpha","text":"new"},{"key":"beta","text":""}]`, invalid: true},
		{name: "malformed JSON", status: 200, input: `[{`},
		{name: "trailing second value", status: 200, input: `[] []`},
		{name: "trailing junk", status: 200, input: `[] garbage`},
		{name: "object", status: 200, input: `{"key":"alpha","text":"new"}`},
		{name: "null", status: 200, input: `null`},
		{name: "partial read", status: 200, readErr: readErr},
		{name: "fetch", fetchErr: fetchErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			path := filepath.Join(root, "snapshot")
			prior := "exact prior bytes \x00 with formatting\n"
			if err := os.WriteFile(path, []byte(prior), 0600); err != nil {
				t.Fatal(err)
			}
			var reader io.Reader = strings.NewReader(tc.input)
			if tc.readErr != nil {
				reader = brokenReader{err: tc.readErr}
			}
			body := &observedBody{reader: reader}
			requests := 0
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				requests++
				if tc.fetchErr != nil {
					return nil, tc.fetchErr
				}
				return &http.Response{StatusCode: tc.status, Body: body, Header: make(http.Header)}, nil
			})}
			err := indexer.Refresh(context.Background(), client, "http://example.test/source", path)
			if err == nil {
				t.Fatal("accepted rejected response")
			}
			for _, want := range []error{tc.readErr, tc.fetchErr} {
				if want != nil && !errors.Is(err, want) {
					t.Fatalf("error = %v; missing %v", err, want)
				}
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("error = %v; want ErrInvalidRecord", err)
			}
			assertFile(t, path, prior)
			wantCloses := 1
			if tc.fetchErr != nil {
				wantCloses = 0
			}
			if requests != 1 || body.closes != wantCloses {
				t.Fatalf("requests/closes = %d/%d; want 1/%d", requests, body.closes, wantCloses)
			}
			assertNoTemps(t, root)
		})
	}
}

func TestRefreshPublicationFailureClosesBody(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "snapshot")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "prior"), []byte("retained"), 0600); err != nil {
		t.Fatal(err)
	}
	body := &observedBody{reader: strings.NewReader(`[]`)}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: body}, nil
	})}
	if err := indexer.Refresh(context.Background(), client, "http://example.test", path); err == nil {
		t.Fatal("expected publication error")
	}
	assertFile(t, filepath.Join(path, "prior"), "retained")
	if body.closes != 1 {
		t.Fatalf("body closes = %d", body.closes)
	}
	assertNoTemps(t, root)
}

func TestRefreshRejectsRedirectWithoutSecondRequest(t *testing.T) {
	t.Parallel()
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		io.WriteString(w, `[]`)
	}))
	t.Cleanup(server.Close)
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: time.Second}
	if err := indexer.Refresh(context.Background(), client, server.URL+"/redirect", path); err == nil {
		t.Fatal("redirect accepted")
	}
	if requests.Load() != 1 {
		t.Fatalf("requests = %d; want 1", requests.Load())
	}
	if client.CheckRedirect != nil {
		t.Fatal("caller client mutated")
	}
	assertFile(t, path, "prior")
}

func TestRefreshRealRequestCancellation(t *testing.T) {
	t.Parallel()
	started, stopped := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		close(stopped)
	}))
	t.Cleanup(server.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	t.Cleanup(cancel)
	done := make(chan error, 1)
	finished := make(chan struct{})
	path := filepath.Join(t.TempDir(), "snapshot")
	if err := os.WriteFile(path, []byte("prior"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cancel(); awaitEvent(t, finished, "HTTP refresh cleanup join") })
	go func() {
		defer close(finished)
		done <- indexer.Refresh(ctx, server.Client(), server.URL, path)
	}()
	awaitEvent(t, started, "HTTP request startup")
	cancel()
	err := awaitResult(t, done)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Refresh cancellation = %v", err)
	}
	awaitEvent(t, stopped, "server request cancellation")
	assertFile(t, path, "prior")
}

func TestRefreshPOSIXOpenReaderKeepsOldSnapshot(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "snapshot")
	old := "complete old snapshot\n"
	if err := os.WriteFile(path, []byte(old), 0600); err != nil {
		t.Fatal(err)
	}
	reader, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { reader.Close() })
	body := &observedBody{reader: strings.NewReader(`[{"key":"alpha","text":"new snapshot"}]`)}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return &http.Response{StatusCode: 200, Body: body}, nil })}
	if err := indexer.Refresh(context.Background(), client, "http://example.test", path); err != nil {
		t.Fatal(err)
	}
	retained, err := io.ReadAll(reader)
	if err != nil || string(retained) != old {
		t.Fatalf("old reader = %q, %v; want %q", retained, err, old)
	}
	assertFile(t, path, "[{\"key\":\"alpha\",\"text\":\"new snapshot\"}]\n")
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("file %s = %q, %v; want %q", path, got, err, want)
	}
}
