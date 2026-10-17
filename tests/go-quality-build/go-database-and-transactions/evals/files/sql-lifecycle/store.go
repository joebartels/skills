package reportworker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const ledgerSQL = "SELECT amount FROM ledger WHERE account = ? ORDER BY sequence"
const currencySQL = "SELECT currency FROM accounts WHERE account = ?"
const deleteSQL = "DELETE FROM reports WHERE account = ?"
const insertSQL = "INSERT INTO reports (account, currency, entry_count, total) VALUES (?, ?, ?, ?)"

type Report struct {
	Account, Currency string
	Entries           int
	Total             int64
}
type Store struct{ db *sql.DB }

func New(db *sql.DB) *Store { return &Store{db: db} }
func (s *Store) Materialize(ctx context.Context, account string, maxEntries int) (Report, error) {
	if maxEntries < 1 {
		return Report{}, errors.New("maxEntries must be positive")
	}
	amounts, err := s.sample(ctx, account, maxEntries)
	if err != nil {
		return Report{}, err
	}
	var currency string
	_ = s.db.QueryRowContext(ctx, currencySQL, account).Scan(&currency)
	report := Report{Account: account, Currency: currency, Entries: len(amounts)}
	for _, amount := range amounts {
		report.Total += amount
	}
	if _, err := s.db.ExecContext(ctx, deleteSQL, account); err != nil {
		return Report{}, fmt.Errorf("remove report: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, insertSQL, account, report.Currency, report.Entries, report.Total); err != nil {
		return Report{}, fmt.Errorf("save report: %w", err)
	}
	return report, nil
}

func (s *Store) sample(ctx context.Context, account string, limit int) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx, ledgerSQL, account)
	if err != nil {
		return nil, fmt.Errorf("read ledger: %w", err)
	}
	defer rows.Close()
	var amounts []int64
	for len(amounts) < limit && rows.Next() {
		var amount int64
		if err := rows.Scan(&amount); err != nil {
			return nil, fmt.Errorf("read amount: %w", err)
		}
		amounts = append(amounts, amount)
	}
	return amounts, nil
}
