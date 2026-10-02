package process

import (
	"context"
	"errors"
	"testing"
)

type sliceError []string

func (e sliceError) Error() string { return "slice cause" }

func hasSliceCause(err error, want sliceError) bool {
	var got sliceError
	return errors.As(err, &got) && len(got) == len(want) && got[0] == want[0]
}

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

func TestCancellationContract(t *testing.T) {
	callbackErr := errors.New("callback failed")
	cause := sliceError{"custom"}

	t.Run("already canceled prevents callbacks, including empty input", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(context.Background())
		cancel(cause)
		called := false
		n, err := Process(ctx, nil, func(context.Context, int) error {
			called = true
			return nil
		})
		if n != 0 || called || !errors.Is(err, context.Canceled) || !hasSliceCause(err, cause) {
			t.Fatalf("accepted=%d, called=%v, err=%v", n, called, err)
		}
	})

	t.Run("empty input succeeds when active", func(t *testing.T) {
		n, err := Process(context.Background(), nil, func(context.Context, int) error { return nil })
		if n != 0 || err != nil {
			t.Fatalf("accepted=%d, err=%v", n, err)
		}
	})

	t.Run("between jobs returns accepted prefix and cause", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(nil)
		var seen []int
		n, err := Process(ctx, []int{1, 2, 3}, func(_ context.Context, job int) error {
			seen = append(seen, job)
			if job == 1 {
				cancel(cause)
			}
			return nil
		})
		if n != 1 || len(seen) != 1 || seen[0] != 1 || !errors.Is(err, context.Canceled) || !hasSliceCause(err, cause) {
			t.Fatalf("accepted=%d, seen=%v, err=%v", n, seen, err)
		}
	})

	t.Run("callback failure retains cancellation and independent error", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(context.Background())
		defer cancel(nil)
		n, err := Process(ctx, []int{1, 2}, func(_ context.Context, _ int) error {
			cancel(cause)
			return callbackErr
		})
		if n != 0 || !errors.Is(err, callbackErr) || !errors.Is(err, context.Canceled) || !hasSliceCause(err, cause) {
			t.Fatalf("accepted=%d, err=%v", n, err)
		}
	})

	t.Run("successful final callback wins coincident cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		n, err := Process(ctx, []int{1}, func(context.Context, int) error {
			cancel()
			return nil
		})
		if n != 1 || err != nil {
			t.Fatalf("accepted=%d, err=%v", n, err)
		}
	})
}
