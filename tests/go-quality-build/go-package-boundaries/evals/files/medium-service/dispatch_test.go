package dispatch

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

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
