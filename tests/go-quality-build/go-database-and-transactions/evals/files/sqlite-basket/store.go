package warehouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrInsufficient = errors.New("insufficient stock")
	ErrInvalid      = errors.New("invalid reservation")
)

type Line struct {
	SKU      string
	Quantity int
}

type Request struct {
	ID         string
	CustomerID string
	Lines      []Line
}

type Receipt struct {
	ID        string
	Units     int
	Remaining map[string]int
}

// Store uses a caller-owned connection pool. Construct it with New.
type Store struct{ db *sql.DB }

func New(db *sql.DB) *Store { return &Store{db: db} }

// Available returns the currently available units, or an error wrapping
// sql.ErrNoRows when the SKU does not exist.
func (s *Store) Available(ctx context.Context, sku string) (int, error) {
	var available int
	if err := s.db.QueryRowContext(ctx, "SELECT available FROM stock WHERE sku=?", sku).Scan(&available); err != nil {
		return 0, fmt.Errorf("read stock %q: %w", sku, err)
	}
	return available, nil
}

// Reserve implements the original single-line worker workflow. Its caller
// supplies an idempotency key in ID; reuse is rejected by the database.
func (s *Store) Reserve(ctx context.Context, req Request) (Receipt, error) {
	units, err := validate(req)
	if err != nil {
		return Receipt{}, err
	}
	if _, err := s.db.ExecContext(ctx, "INSERT INTO reservations(id,customer_id,units) VALUES (?,?,?)", req.ID, req.CustomerID, units); err != nil {
		return Receipt{}, fmt.Errorf("record reservation: %w", err)
	}
	receipt := Receipt{ID: req.ID, Units: units, Remaining: make(map[string]int)}
	for _, line := range req.Lines {
		available, err := s.Available(ctx, line.SKU)
		if err != nil {
			return Receipt{}, err
		}
		if available < line.Quantity {
			return Receipt{}, fmt.Errorf("%s: %w", line.SKU, ErrInsufficient)
		}
		remaining := available - line.Quantity
		if _, err := s.db.ExecContext(ctx, "UPDATE stock SET available=? WHERE sku=?", remaining, line.SKU); err != nil {
			return Receipt{}, fmt.Errorf("update stock: %w", err)
		}
		if _, err := s.db.ExecContext(ctx, "INSERT INTO reservation_lines(reservation_id,sku,quantity) VALUES (?,?,?)", req.ID, line.SKU, line.Quantity); err != nil {
			return Receipt{}, fmt.Errorf("record reservation line: %w", err)
		}
		receipt.Remaining[line.SKU] = remaining
	}
	return receipt, nil
}

func validate(req Request) (int, error) {
	if req.ID == "" || req.CustomerID == "" || len(req.Lines) == 0 {
		return 0, ErrInvalid
	}
	seen := make(map[string]bool)
	units := 0
	for _, line := range req.Lines {
		if line.SKU == "" || line.Quantity <= 0 || seen[line.SKU] {
			return 0, ErrInvalid
		}
		seen[line.SKU] = true
		units += line.Quantity
	}
	return units, nil
}
