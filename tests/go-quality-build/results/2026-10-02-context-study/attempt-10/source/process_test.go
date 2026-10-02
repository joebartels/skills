package process

import (
	"context"
	"errors"
	"testing"
)

type sliceError struct{ values []int }

func (e sliceError) Error() string { return "custom cancellation cause" }

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

func TestProcessCancellationContracts(t *testing.T) {
	t.Run("already canceled skips callbacks and preserves cause", func(t *testing.T) {
		cause := errors.New("stop now")
		ctx, cancel := context.WithCancelCause(context.Background())
		cancel(cause)
		calls := 0
		n, err := Process(ctx, []int{1}, func(context.Context, int) error {
			calls++
			return nil
		})
		if n != 0 || calls != 0 || !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
			t.Fatalf("accepted=%d calls=%d err=%v", n, calls, err)
		}
	})

	t.Run("empty live input succeeds", func(t *testing.T) {
		n, err := Process(context.Background(), nil, func(context.Context, int) error {
			t.Fatal("callback invoked for empty input")
			return nil
		})
		if n != 0 || err != nil {
			t.Fatalf("accepted=%d err=%v, want 0, nil", n, err)
		}
	})

	t.Run("cancellation between jobs returns accepted prefix", func(t *testing.T) {
		cause := errors.New("pause")
		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(nil)
		calls := 0
		n, err := Process(ctx, []int{2, 4, 6}, func(context.Context, int) error {
			calls++
			if calls == 1 {
				cancel(cause)
			}
			return nil
		})
		if n != 1 || calls != 1 || !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
			t.Fatalf("accepted=%d calls=%d err=%v", n, calls, err)
		}
	})

	t.Run("callback failure and non-comparable cancellation cause are retained", func(t *testing.T) {
		callbackErr := errors.New("callback failed")
		cause := sliceError{values: []int{1, 2}}
		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(nil)
		n, err := Process(ctx, []int{1, 2}, func(context.Context, int) error {
			cancel(cause)
			return callbackErr
		})
		var gotCause sliceError
		if n != 0 || !errors.Is(err, callbackErr) || !errors.Is(err, context.Canceled) || !errors.As(err, &gotCause) || len(gotCause.values) != 2 {
			t.Fatalf("accepted=%d err=%v cause=%v", n, err, gotCause)
		}
	})

	t.Run("successful last callback wins coincident cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		n, err := Process(ctx, []int{9}, func(context.Context, int) error {
			cancel()
			return nil
		})
		if n != 1 || err != nil {
			t.Fatalf("accepted=%d err=%v, want 1, nil", n, err)
		}
	})
}
