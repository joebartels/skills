package billing

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"example.com/commerce/internal/invoice"
)

type gatewayFunc func(context.Context, string, int) error

func (f gatewayFunc) Charge(ctx context.Context, id string, cents int) error {
	return f(ctx, id, cents)
}

func TestChargePersistsCanonicalRowAfterSuccess(t *testing.T) {
	var ledger bytes.Buffer
	var gotID string
	var gotCents int
	s := Service{Ledger: &ledger, Gateway: gatewayFunc(func(_ context.Context, id string, cents int) error {
		gotID, gotCents = id, cents
		return nil
	})}
	if err := s.Charge(context.Background(), "invoice-1", 250); err != nil {
		t.Fatal(err)
	}
	if gotID != "invoice-1" || gotCents != 250 || ledger.String() != "invoice-1,250\n" {
		t.Fatalf("gateway got (%q, %d), ledger %q", gotID, gotCents, ledger.String())
	}
}

func TestChargeFailureDoesNotWrite(t *testing.T) {
	var ledger bytes.Buffer
	failure := errors.New("declined")
	s := Service{Ledger: &ledger, Gateway: gatewayFunc(func(context.Context, string, int) error { return failure })}
	if err := s.Charge(context.Background(), "invoice-1", 250); !errors.Is(err, failure) || ledger.Len() != 0 {
		t.Fatalf("error = %v, ledger %q", err, ledger.String())
	}
}

func TestChargeRejectsInvalidBeforeGateway(t *testing.T) {
	called := false
	s := Service{Ledger: &bytes.Buffer{}, Gateway: gatewayFunc(func(context.Context, string, int) error { called = true; return nil })}
	if err := s.Charge(context.Background(), "bad,id", 250); err == nil || called {
		t.Fatalf("error = %v, gateway called = %v", err, called)
	}
}

func TestReadLedgerDecodesRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger")
	if err := os.WriteFile(path, []byte("invoice-1,250\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var rows []invoice.Row
	if err := ReadLedger(path, func(row invoice.Row) error { rows = append(rows, row); return nil }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rows, []invoice.Row{{ID: "invoice-1", Cents: 250}}) {
		t.Fatalf("rows = %#v", rows)
	}
}
