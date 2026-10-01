package billing

import (
	"context"
	"fmt"
	"io"
	"os"

	"example.com/commerce/internal/invoice"
)

// Gateway is the billing-facing contract; transport details live in an adapter.
type Gateway interface {
	Charge(context.Context, string, int) error
}

// Service applies billing rules and appends successful charges to the ledger.
type Service struct {
	Ledger  io.Writer
	Gateway Gateway
}

func (s *Service) Charge(ctx context.Context, id string, cents int) error {
	encoded, err := invoice.Encode(invoice.Row{ID: id, Cents: cents})
	if err != nil {
		return err
	}
	if s.Gateway == nil {
		return fmt.Errorf("billing gateway is required")
	}
	if err := s.Gateway.Charge(ctx, id, cents); err != nil {
		return err
	}
	_, err = s.Ledger.Write(encoded)
	return err
}

// Export copies the existing ledger bytes to out.
// ReadLedger decodes stored ledger rows using the same codec as settlement input.
func ReadLedger(path string, apply func(invoice.Row) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return invoice.Read(f, apply)
}

func Export(path string, out io.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(out, f)
	return err
}
