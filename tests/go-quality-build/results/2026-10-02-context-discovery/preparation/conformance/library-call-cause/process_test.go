package process

import (
	"context"
	"errors"
	"testing"
)

func TestSequentialProgress(t *testing.T) {
	failure := errors.New("operation failed")
	var seen []int
	n, err := Process(context.Background(), []int{3, 5, 7}, func(_ context.Context, job int) error {
		seen = append(seen, job)
		if job == 5 {
			return failure
		}
		return nil
	})
	if n != 1 || !errors.Is(err, failure) || len(seen) != 2 || seen[0] != 3 || seen[1] != 5 {
		t.Fatalf("accepted=%d, err=%v, seen=%v", n, err, seen)
	}
}
