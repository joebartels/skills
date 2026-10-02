package metadata_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"example.com/metadata"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFetchMetadata(t *testing.T) {
	client := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"name":"first","revision":1}`))}, nil
	})}
	got, err := metadata.Fetch(context.Background(), client, "http://example.invalid/catalog")
	if err != nil || got.Name != "first" || got.Revision != 1 {
		t.Fatalf("Fetch = %#v, %v", got, err)
	}
}
