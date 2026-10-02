package metadata_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"example.com/metadata"
)

type probeTransport func(*http.Request) (*http.Response, error)

func (f probeTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type probeBody struct {
	io.Reader
	closed atomic.Bool
}

func (b *probeBody) Close() error { b.closed.Store(true); return nil }

func TestHTTPCancelAndBodyLifetime(t *testing.T) {
	for _, status := range []int{200, 503} {
		body := &probeBody{Reader: strings.NewReader(`{"name":" exact ","revision":2}`)}
		client := &http.Client{Transport: probeTransport(func(r *http.Request) (*http.Response, error) {
			if r.Method != "GET" || r.URL.RequestURI() != "/catalog?q=2" {
				t.Errorf("request = %s %s", r.Method, r.URL.RequestURI())
			}
			return &http.Response{StatusCode: status, Body: body}, nil
		})}
		got, err := metadata.Fetch(context.Background(), client, "http://example.invalid/catalog?q=2")
		if status == 200 && (err != nil || got != (metadata.Metadata{Name: " exact ", Revision: 2})) {
			t.Fatalf("success = %#v, %v", got, err)
		}
		if status != 200 && (err == nil || got != (metadata.Metadata{})) {
			t.Fatalf("failure = %#v, %v", got, err)
		}
		if !body.closed.Load() {
			t.Fatal("response body was not closed")
		}
	}
	started := make(chan struct{})
	canceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
		close(canceled)
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client := server.Client()
	client.Timeout = 2 * time.Second
	done := make(chan error, 1)
	go func() {
		_, err := metadata.Fetch(ctx, client, server.URL+"/catalog")
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("HTTP request did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Fetch did not finish after caller cancellation")
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("server did not observe request cancellation")
	}
}
