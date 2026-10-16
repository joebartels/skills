package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExistingSuccess(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) }))
	defer s.Close()
	got, err := New(s.Client(), false).Fetch(context.Background(), s.URL)
	if err != nil || string(got) != "ok" {
		t.Fatalf("got %q, %v", got, err)
	}
}
