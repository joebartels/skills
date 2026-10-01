package dispatch

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestReplaySubmitsNonblankTrimmedLinesSequentially(t *testing.T) {
	var notified []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		notified = append(notified, string(b))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()
	s := Service{Dir: t.TempDir(), NotifyURL: ts.URL, Client: ts.Client()}
	if err := s.Replay(context.Background(), bytes.NewBufferString("  a  \n\n b\t\n")); err != nil {
		t.Fatal(err)
	}
	if len(notified) != 2 || notified[0] != "a" || notified[1] != "b" {
		t.Fatalf("notifications: %q", notified)
	}
	for _, id := range []string{"a", "b"} {
		b, err := os.ReadFile(filepath.Join(s.Dir, id))
		if err != nil || string(b) != "queued\n" {
			t.Fatalf("record %q: %q, %v", id, b, err)
		}
	}
}

func TestReplayStopsOnInvalidID(t *testing.T) {
	var calls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()
	s := Service{Dir: t.TempDir(), NotifyURL: ts.URL, Client: ts.Client()}
	err := s.Replay(context.Background(), bytes.NewBufferString("first\n../bad\nlast\n"))
	if err == nil {
		t.Fatal("expected invalid ID error")
	}
	if calls != 1 {
		t.Fatalf("notification calls = %d, want 1", calls)
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "last")); !os.IsNotExist(err) {
		t.Fatalf("later record exists or stat failed: %v", err)
	}
}

func TestReplayStopsOnNotificationFailure(t *testing.T) {
	var calls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()
	s := Service{Dir: t.TempDir(), NotifyURL: ts.URL, Client: ts.Client()}
	err := s.Replay(context.Background(), bytes.NewBufferString("first\nsecond\n"))
	if err == nil {
		t.Fatal("expected notification error")
	}
	if calls != 1 {
		t.Fatalf("notification calls = %d, want 1", calls)
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "first")); err != nil {
		t.Fatalf("first record missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "second")); !os.IsNotExist(err) {
		t.Fatalf("second record exists or stat failed: %v", err)
	}
}

func TestSubmit(t *testing.T) {
	var notified string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		notified = string(b)
		w.WriteHeader(204)
	}))
	defer ts.Close()
	s := Service{Dir: t.TempDir(), NotifyURL: ts.URL, Client: ts.Client()}
	if err := s.Submit(context.Background(), "a"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(s.Dir, "a"))
	if err != nil || string(b) != "queued\n" || notified != "a" {
		t.Fatalf("record %q, notified %q, error %v", b, notified, err)
	}
}
func TestNotificationFailureKeepsRecord(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer ts.Close()
	s := Service{Dir: t.TempDir(), NotifyURL: ts.URL, Client: ts.Client()}
	if s.Submit(context.Background(), "a") == nil {
		t.Fatal("failure reported as success")
	}
	if _, err := os.Stat(filepath.Join(s.Dir, "a")); err != nil {
		t.Fatal(err)
	}
}
