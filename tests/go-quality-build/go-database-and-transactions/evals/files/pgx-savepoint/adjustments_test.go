package ledger

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type acceptingTx struct{ pgx.Tx }

func (t acceptingTx) Begin(context.Context) (pgx.Tx, error) { return t, nil }
func (t acceptingTx) Commit(context.Context) error          { return nil }
func (t acceptingTx) Rollback(context.Context) error        { return nil }

func (t acceptingTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func TestApplyAdjustmentsExistingCases(t *testing.T) {
	if err := ApplyAdjustments(context.Background(), acceptingTx{}, nil); err != nil {
		t.Fatal(err)
	}
	if err := ApplyAdjustments(context.Background(), acceptingTx{}, []Adjustment{{"alice", 12}}); err != nil {
		t.Fatal(err)
	}
	if err := ApplyAdjustments(context.Background(), acceptingTx{}, []Adjustment{{"", 12}}); !errors.Is(err, ErrInvalidAdjustment) {
		t.Fatalf("invalid item: %v", err)
	}
}
