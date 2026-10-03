package finalize

import (
	"context"
	"testing"
	"time"
)

func TestCompletedWorkFinalizes(t *testing.T) {
	finalized := -1
	n, err := Run(context.Background(), []int{1, 2}, func(context.Context, int) error { return nil }, func(_ context.Context, n int) error { finalized = n; return nil }, time.Second)
	if n != 2 || finalized != 2 || err != nil {
		t.Fatalf("accepted=%d finalized=%d err=%v", n, finalized, err)
	}
}
