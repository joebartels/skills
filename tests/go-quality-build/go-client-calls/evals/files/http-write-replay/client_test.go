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
	got, err := Submit(context.Background(), s.Client(), s.URL, "op-1", []byte("x"), true)
	if err != nil || string(got) != "ok" {
		t.Fatalf("got %q, %v", got, err)
	}
}
