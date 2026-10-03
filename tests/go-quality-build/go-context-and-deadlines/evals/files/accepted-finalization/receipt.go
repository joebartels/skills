package receipt

import (
	"context"
	"time"
)

// ApplyBatch applies ordered items and writes a receipt for accepted work.
func ApplyBatch(ctx context.Context, items []string, finalizationBudget time.Duration, apply func(context.Context, string) error, finish func(context.Context, int) error) (int, error) {
	n := 0
	for _, item := range items {
		if err := apply(ctx, item); err != nil {
			return n, err
		}
		n++
	}
	if n == 0 {
		return 0, nil
	}
	return n, finish(ctx, n)
}
