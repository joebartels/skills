package indexer_test

import (
	"context"
	"errors"
	"example.com/indexer"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestControllerAcceptedPrefix(t *testing.T) {
	root := t.TempDir()
	i := indexer.Open(root)
	if err := i.ApplyCSV(strings.NewReader("a,old\na,new\nb,kept\na,  \nc,later\n")); !errors.Is(err, indexer.ErrInvalidRecord) {
		t.Fatalf("invalid row: %v", err)
	}
	for key, want := range map[string]string{"a": "new", "b": "kept"} {
		got, err := i.Get(key)
		if err != nil || got != want {
			t.Fatalf("%s=%q,%v want %q", key, got, err, want)
		}
	}
	if _, err := i.Get("c"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("later row exists: %v", err)
	}
}

type controllerTransport func(*http.Request) (*http.Response, error)

func (f controllerTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestControllerSnapshotPublication(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.json")
	old := `[{"key":"old","text":"Previous"}]` + "\n"
	if err := os.WriteFile(path, []byte(old), 0600); err != nil {
		t.Fatal(err)
	}
	opened, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer opened.Close()
	body := `[{"key":"a","text":"Alpha"},{"key":"b","text":"Beta"}]`
	client := &http.Client{Transport: controllerTransport(func(r *http.Request) (*http.Response, error) {
		if r.Method != "GET" {
			t.Errorf("method %s", r.Method)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	if err := indexer.Refresh(context.Background(), client, "http://example.invalid/index", path); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != body+"\n" {
		t.Fatalf("new snapshot=%q,%v", got, err)
	}
	prior, err := io.ReadAll(opened)
	if err != nil || string(prior) != old {
		t.Fatalf("opened prior snapshot=%q,%v", prior, err)
	}
	body = `[{"key":"a","text":"  "}]`
	if err := indexer.Refresh(context.Background(), client, "http://example.invalid/index", path); !errors.Is(err, indexer.ErrInvalidRecord) {
		t.Fatalf("invalid text=%v", err)
	}
	retained, err := os.ReadFile(path)
	if err != nil || string(retained) != string(got) {
		t.Fatalf("rejection changed snapshot=%q,%v", retained, err)
	}
	body = `[{"key":"a","text":"Alpha"}] true`
	if err := indexer.Refresh(context.Background(), client, "http://example.invalid/index", path); err == nil {
		t.Fatal("trailing JSON accepted")
	}
}
func TestControllerHostJoinAndErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	cleaning := make(chan struct{})
	allow := make(chan struct{})
	finished := make(chan struct{})
	released := make(chan bool, 1)
	result := make(chan error, 1)
	workErr := errors.New("work")
	releaseErr := errors.New("release")
	t.Cleanup(func() {
		cancel()
		select {
		case <-allow:
		default:
			close(allow)
		}
		select {
		case <-result:
		case <-time.After(5 * time.Second):
			t.Error("host did not join")
		}
	})
	go func() {
		result <- indexer.Serve(ctx, 20*time.Millisecond, func(c context.Context) error {
			close(started)
			<-c.Done()
			close(cleaning)
			<-allow
			close(finished)
			return workErr
		}, func() error {
			select {
			case <-finished:
				released <- true
			default:
				released <- false
			}
			return releaseErr
		})
	}()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("callback not started")
	}
	cancel()
	select {
	case <-cleaning:
	case <-time.After(3 * time.Second):
		t.Fatal("caller cancellation absent")
	}
	select {
	case <-released:
		t.Fatal("release before callback cleanup")
	case <-time.After(20 * time.Millisecond):
	}
	close(allow)
	select {
	case err := <-result:
		if !errors.Is(err, workErr) || !errors.Is(err, releaseErr) {
			t.Fatalf("error identities=%v", err)
		}
		result <- nil
	case <-time.After(3 * time.Second):
		t.Fatal("host completion absent")
	}
	if !<-released {
		t.Fatal("release preceded completion")
	}
}
func TestControllerSuccessfulRecurrence(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entered := make(chan int, 3)
	gate := make(chan struct{})
	completed := make(chan time.Time, 1)
	second := make(chan time.Time, 1)
	result := make(chan error, 1)
	calls := 0
	t.Cleanup(func() {
		cancel()
		select {
		case <-gate:
		default:
			close(gate)
		}
		select {
		case <-result:
		case <-time.After(5 * time.Second):
			t.Error("recurrence host did not join")
		}
	})
	go func() {
		result <- indexer.Serve(ctx, 50*time.Millisecond, func(context.Context) error {
			calls++
			entered <- calls
			if calls == 1 {
				<-gate
				completed <- time.Now()
			} else {
				second <- time.Now()
				cancel()
			}
			return nil
		}, func() error { return nil })
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("no initial refresh")
	}
	time.Sleep(70 * time.Millisecond)
	close(gate)
	finish := <-completed
	select {
	case next := <-second:
		if gap := next.Sub(finish); gap < 45*time.Millisecond {
			t.Fatalf("recurrence gap %v < completion-relative interval", gap)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no successful recurrence")
	}
}
func TestControllerActualProcessFailures(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "indexer")
	build := exec.Command("go", "build", "-o", binary, "./cmd/indexer")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "apply", "--dir", dir)
	cmd.Stdin = strings.NewReader("a,first\nb,  \nc,later\n")
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("apply status/streams=%v/%q/%q", err, stdout.String(), stderr.String())
	}
	got, err := os.ReadFile(filepath.Join(dir, "a"))
	if err != nil || string(got) != "first" {
		t.Fatalf("process prefix=%q,%v", got, err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer server.Close()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	cmd = exec.CommandContext(ctx2, binary, "watch", "--url", server.URL, "--file", filepath.Join(t.TempDir(), "index.json"), "--interval", "10ms")
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if !errors.As(err, &exit) || exit.ExitCode() != 2 || stdout.Len() != 0 || stderr.Len() == 0 {
		t.Fatalf("watch failed-start status/streams=%v/%q/%q", err, stdout.String(), stderr.String())
	}
}
