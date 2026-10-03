package stages

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestSequentialBodies(t *testing.T) {
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(r.URL.Path)), Header: make(http.Header), Request: r}, nil
	})}
	got, err := FetchAll(context.Background(), client, []string{"http://example.test/first", "http://example.test/second"}, time.Second, time.Second)
	if err != nil || len(got) != 2 || string(got[0]) != "/first" || string(got[1]) != "/second" {
		t.Fatalf("bodies=%q err=%v", got, err)
	}
}
