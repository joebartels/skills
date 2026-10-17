package ledger

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidAdjustment = errors.New("invalid adjustment")
	ErrAccountMissing    = errors.New("account missing")
)

type Adjustment struct {
	AccountID string
	Delta     int64
}

// ApplyAdjustments stages account adjustments in the caller's transaction.
func ApplyAdjustments(ctx context.Context, parent pgx.Tx, items []Adjustment) error {
	for i, item := range items {
		if item.AccountID == "" {
			return fmt.Errorf("adjustment %d: %w", i, ErrInvalidAdjustment)
		}
		tag, err := parent.Exec(ctx,
			"update balances set cents = cents + $1 where account_id = $2",
			item.Delta, item.AccountID)
		if err != nil {
			return fmt.Errorf("adjustment %d: %w", i, err)
		}
		if tag.RowsAffected() != 1 {
			return fmt.Errorf("adjustment %d: %w", i, ErrAccountMissing)
		}
	}
	return nil
}
