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
	"sync"
	"testing"
	"time"

	"example.com/indexer"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackedBody struct {
	io.Reader
	closes int
}

func (b *trackedBody) Close() error { b.closes++; return nil }

func TestRefreshPublishesOrderedSnapshot(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("open-reader replacement snapshot is a POSIX execution contract")
	}
	path := filepath.Join(t.TempDir(), "snapshot.json")
	old := "exact old snapshot\x00\n"
	writeFixture(t, path, old)
	before, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { before.Close() })
	body := &trackedBody{Reader: strings.NewReader(" [ {\"key\":\"beta\",\"text\":\"  keep\\ntext  \"}, {\"key\":\"alpha\",\"text\":\"first\"}, {\"key\":\"alpha\",\"text\":\"second\"} ] \n\t")}
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "caller")
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != http.MethodGet || r.URL.String() != "http://fixture.invalid/records?x=1" || r.Context().Value(contextKey{}) != "caller" {
			return nil, errors.New("wrong request or caller context")
		}
		return &http.Response{StatusCode: 200, Body: body}, nil
	})}
	if err := indexer.Refresh(ctx, client, "http://fixture.invalid/records?x=1", path); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || body.closes != 1 {
		t.Fatalf("requests=%d, closes=%d; want one each", calls, body.closes)
	}
	assertFile(t, path, "[{\"key\":\"beta\",\"text\":\"  keep\\ntext  \"},{\"key\":\"alpha\",\"text\":\"first\"},{\"key\":\"alpha\",\"text\":\"second\"}]\n")
	gotOld, err := io.ReadAll(before)
	if err != nil || string(gotOld) != old {
		t.Fatalf("opened snapshot = %q, %v; want %q", gotOld, err, old)
	}
	assertNoSnapshotTemps(t, filepath.Dir(path))
}

func TestRefreshEmptyArray(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "new.json")
	body := &trackedBody{Reader: strings.NewReader("[]")}
	client := responseClient(200, body)
	if err := indexer.Refresh(context.Background(), client, "http://fixture.invalid", path); err != nil {
		t.Fatal(err)
	}
	assertFile(t, path, "[]\n")
	if body.closes != 1 {
		t.Fatalf("closes=%d; want 1", body.closes)
	}
}

func TestRefreshRejectionKeepsExactPriorBytes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		status  int
		payload string
		invalid bool
	}{
		{"status", 201, "[]", false},
		{"empty", 200, "", false},
		{"null", 200, "null", false},
		{"object", 200, "{}", false},
		{"malformed", 200, "[{", false},
		{"trailing-object", 200, "[] {}", false},
		{"trailing-garbage", 200, "[] garbage", false},
		{"bad-key", 200, `[{"key":"Alpha","text":"valid"}]`, true},
		{"path-key", 200, `[{"key":"../escape","text":"valid"}]`, true},
		{"blank", 200, `[{"key":"alpha","text":" \n\t\u2003"}]`, true},
		{"later-invalid", 200, `[{"key":"alpha","text":"valid"},{"key":"beta","text":""}]`, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "snapshot.json")
			old := "old bytes are deliberately not JSON\x00\n"
			writeFixture(t, path, old)
			body := &trackedBody{Reader: strings.NewReader(test.payload)}
			err := indexer.Refresh(context.Background(), responseClient(test.status, body), "http://fixture.invalid", path)
			if err == nil || (test.invalid && !errors.Is(err, indexer.ErrInvalidRecord)) {
				t.Fatalf("Refresh = %v; want rejection (invalid=%v)", err, test.invalid)
			}
			if body.closes != 1 {
				t.Fatalf("closes=%d; want 1", body.closes)
			}
			assertFile(t, path, old)
			assertNoSnapshotTemps(t, filepath.Dir(path))
		})
	}
}

func TestRefreshReadAndFetchErrors(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"read", "fetch"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "snapshot.json")
			writeFixture(t, path, "prior")
			failure := errors.New(name + " failure")
			body := &trackedBody{Reader: io.MultiReader(strings.NewReader("[]"), errorReader{failure})}
			client := responseClient(200, body)
			if name == "fetch" {
				client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, failure })
			}
			if err := indexer.Refresh(context.Background(), client, "http://fixture.invalid", path); !errors.Is(err, failure) {
				t.Fatalf("Refresh = %v; want failure identity", err)
			}
			wantCloses := 1
			if name == "fetch" {
				wantCloses = 0
			}
			if body.closes != wantCloses {
				t.Fatalf("closes=%d; want %d", body.closes, wantCloses)
			}
			assertFile(t, path, "prior")
		})
	}
}

func TestRefreshPublicationFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "destination")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(path, "prior")
	writeFixture(t, marker, "retained directory contents")
	body := &trackedBody{Reader: strings.NewReader("[]")}
	if err := indexer.Refresh(context.Background(), responseClient(200, body), "http://fixture.invalid", path); err == nil {
		t.Fatal("wanted rename failure")
	}
	assertFile(t, marker, "retained directory contents")
	if body.closes != 1 {
		t.Fatalf("closes=%d; want 1", body.closes)
	}
	assertNoSnapshotTemps(t, root)
}

func TestRefreshPublicationFailureKeepsPriorFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "snapshot.json")
	writeFixture(t, path, "prior exact bytes\x00\n")
	if err := os.Chmod(root, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(root, 0700) })
	probe, err := os.CreateTemp(root, ".permission-probe-*")
	if err == nil {
		probe.Close()
		os.Remove(probe.Name())
		t.Skip("filesystem privileges bypass directory write permission")
	}
	body := &trackedBody{Reader: strings.NewReader("[]")}
	if err := indexer.Refresh(context.Background(), responseClient(200, body), "http://fixture.invalid", path); err == nil {
		t.Fatal("wanted publication failure")
	}
	assertFile(t, path, "prior exact bytes\x00\n")
	if body.closes != 1 {
		t.Fatalf("closes=%d; want 1", body.closes)
	}
}

func TestRefreshRealHTTPCancellation(t *testing.T) {
	t.Parallel()
	started := make(chan struct{})
	handlerDone := make(chan struct{})
	unblockHandler := make(chan struct{})
	var unblockOnce sync.Once
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		defer close(handlerDone)
		select {
		case <-r.Context().Done():
		case <-unblockHandler:
		}
	}))
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	finished := make(chan struct{})
	path := filepath.Join(t.TempDir(), "snapshot.json")
	writeFixture(t, path, "prior")
	t.Cleanup(func() {
		cancel()
		unblockOnce.Do(func() { close(unblockHandler) })
		awaitEvent(t, finished, "refresh completion during cleanup")
		server.Close()
	})
	go func() { defer close(finished); result <- indexer.Refresh(ctx, server.Client(), server.URL, path) }()
	awaitEvent(t, started, "HTTP request arrival")
	cancel()
	err := awaitError(t, result, "canceled refresh")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Refresh = %v; want cancellation", err)
	}
	awaitEvent(t, handlerDone, "server request cancellation")
	assertFile(t, path, "prior")
}

func responseClient(status int, body io.ReadCloser) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Body: body}, nil
	})}
}
func writeFixture(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}
func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("ReadFile = %q, %v; want %q", got, err, want)
	}
}
func assertNoSnapshotTemps(t *testing.T, root string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, ".snapshot-*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("temporary snapshots = %v, %v; want none", matches, err)
	}
}
func awaitEvent(t *testing.T, event <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-event:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}
func awaitError(t *testing.T, result <-chan error, what string) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
		return nil
	}
}
