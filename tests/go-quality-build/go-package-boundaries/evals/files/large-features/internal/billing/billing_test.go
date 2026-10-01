package billing

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCharge(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	defer ts.Close()
	var ledger bytes.Buffer
	s := Service{Ledger: &ledger, Client: ts.Client(), GatewayURL: ts.URL}
	if err := s.Charge(context.Background(), "invoice-1", 250); err != nil {
		t.Fatal(err)
	}
	if ledger.String() != "invoice-1,250\n" {
		t.Fatalf("got %q", ledger.String())
	}
}
func TestDeclineDoesNotWrite(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(402) }))
	defer ts.Close()
	var ledger bytes.Buffer
	s := Service{Ledger: &ledger, Client: ts.Client(), GatewayURL: ts.URL}
	if s.Charge(context.Background(), "invoice-1", 250) == nil || ledger.Len() != 0 {
		t.Fatal("declined charge persisted or reported successful")
	}
}
