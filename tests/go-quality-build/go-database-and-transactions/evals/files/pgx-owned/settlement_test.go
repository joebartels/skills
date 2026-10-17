package settlement

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type postingTx struct {
	pgx.Tx
	calls   int
	failAt  int
	err     error
	amounts []int64
}

func (tx *postingTx) Exec(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	tx.calls++
	if tx.calls == tx.failAt {
		return pgconn.CommandTag{}, tx.err
	}
	tx.amounts = append(tx.amounts, args[0].(int64))
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func TestPostStagesInInputOrder(t *testing.T) {
	tx := &postingTx{}
	n, err := Post(context.Background(), tx, []Item{{1, 9}, {2, -4}})
	if n != 2 || err != nil || len(tx.amounts) != 2 || tx.amounts[0] != 9 || tx.amounts[1] != -4 {
		t.Fatalf("Post = %d, %v, amounts %v", n, err, tx.amounts)
	}
}

func TestPostReportsPrefixAndPreservesFailure(t *testing.T) {
	want := errors.New("write failed")
	tx := &postingTx{failAt: 2, err: want}
	n, err := Post(context.Background(), tx, []Item{{1, 9}, {2, -4}, {3, 8}})
	if n != 1 || !errors.Is(err, want) || tx.calls != 2 {
		t.Fatalf("Post = %d, %v; calls %d", n, err, tx.calls)
	}
}

func TestPostEmptyDoesNotTouchBorrowedTransaction(t *testing.T) {
	if n, err := Post(context.Background(), nil, nil); n != 0 || err != nil {
		t.Fatalf("Post = %d, %v", n, err)
	}
}
