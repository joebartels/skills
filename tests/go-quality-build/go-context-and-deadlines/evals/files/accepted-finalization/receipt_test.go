package receipt

import (
	"context"
	"testing"
	"time"
)

func TestReceiptSuccess(t *testing.T) {
	var accepted []string
	n, err := ApplyBatch(context.Background(), []string{"one", "two"}, time.Second,
		func(_ context.Context, item string) error { accepted = append(accepted, item); return nil },
		func(_ context.Context, count int) error {
			if count != 2 {
				t.Errorf("receipt count=%d", count)
			}
			return nil
		})
	if n != 2 || err != nil || len(accepted) != 2 || accepted[0] != "one" || accepted[1] != "two" {
		t.Fatalf("count=%d accepted=%v error=%v", n, accepted, err)
	}
}
