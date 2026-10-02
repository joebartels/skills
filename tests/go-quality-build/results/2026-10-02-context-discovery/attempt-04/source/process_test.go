package process

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type sliceCause struct{ values []int }

func (sliceCause) Error() string { return "custom cancellation cause" }

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
	if n != 1 || !errors.Is(err, failure) || !reflect.DeepEqual(seen, []int{3, 5}) {
		t.Fatalf("accepted=%d, err=%v, seen=%v", n, err, seen)
	}
}

func TestProcessCancellationAndResults(t *testing.T) {
	t.Run("already canceled", func(t *testing.T) {
		cause := sliceCause{values: []int{1, 2}}
		ctx, cancel := context.WithCancelCause(context.Background())
		cancel(cause)
		calls := 0
		n, err := Process(ctx, []int{1}, func(context.Context, int) error { calls++; return nil })
		var gotCause sliceCause
		if n != 0 || calls != 0 || !errors.Is(err, context.Canceled) || !errors.As(err, &gotCause) || !reflect.DeepEqual(gotCause, cause) {
			t.Fatalf("accepted=%d, calls=%d, err=%v, cause=%v", n, calls, err, gotCause)
		}
	})

	t.Run("empty input succeeds", func(t *testing.T) {
		n, err := Process(context.Background(), nil, func(context.Context, int) error { t.Fatal("callback called"); return nil })
		if n != 0 || err != nil {
			t.Fatalf("accepted=%d, err=%v", n, err)
		}
	})

	t.Run("between jobs preserves accepted prefix and cause", func(t *testing.T) {
		cause := sliceCause{values: []int{4}}
		ctx, cancel := context.WithCancelCause(context.Background())
		var seen []int
		n, err := Process(ctx, []int{2, 4, 6}, func(_ context.Context, job int) error {
			seen = append(seen, job)
			if job == 2 {
				cancel(cause)
			}
			return nil
		})
		var gotCause sliceCause
		if n != 1 || !reflect.DeepEqual(seen, []int{2}) || !errors.Is(err, context.Canceled) || !errors.As(err, &gotCause) || !reflect.DeepEqual(gotCause, cause) {
			t.Fatalf("accepted=%d, seen=%v, err=%v, cause=%v", n, seen, err, gotCause)
		}
	})

	t.Run("last accepted callback wins coincident cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		n, err := Process(ctx, []int{1}, func(context.Context, int) error { cancel(); return nil })
		if n != 1 || err != nil {
			t.Fatalf("accepted=%d, err=%v", n, err)
		}
	})

	t.Run("callback and observed cancellation errors are retained", func(t *testing.T) {
		cause := sliceCause{values: []int{7, 8}}
		failure := errors.New("callback failed")
		ctx, cancel := context.WithCancelCause(context.Background())
		n, err := Process(ctx, []int{9, 10}, func(context.Context, int) error {
			cancel(cause)
			return failure
		})
		var gotCause sliceCause
		if n != 0 || !errors.Is(err, failure) || !errors.Is(err, context.Canceled) || !errors.As(err, &gotCause) || !reflect.DeepEqual(gotCause, cause) {
			t.Fatalf("accepted=%d, err=%v, cause=%v", n, err, gotCause)
		}
	})
}
