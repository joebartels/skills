// Package settlement posts pending account adjustments.
package settlement

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Beginner is supplied by the host, normally as a *pgxpool.Pool.
// The host retains ownership of it.
type Beginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

var _ Beginner = (*pgxpool.Pool)(nil)

type Store struct{ db Beginner }

func New(db Beginner) *Store { return &Store{db: db} }

type Item struct {
	AccountID   int64
	AmountCents int64
}

type Report struct {
	RunID   int64
	Applied int
}

var ErrAccountMissing = errors.New("account missing")

// Post stages adjustments in order. It returns the number staged before a
// failure. The caller owns tx, including committing or rolling it back.
func Post(ctx context.Context, tx pgx.Tx, items []Item) (int, error) {
	for i, item := range items {
		tag, err := tx.Exec(ctx,
			`UPDATE accounts SET balance_cents = balance_cents + $1 WHERE id = $2`,
			item.AmountCents, item.AccountID)
		if err != nil {
			return i, fmt.Errorf("post account %d: %w", item.AccountID, err)
		}
		if tag.RowsAffected() != 1 {
			return i, fmt.Errorf("post account %d: %w", item.AccountID, ErrAccountMissing)
		}
	}
	return len(items), nil
}
