package finalize

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunFinalizationContract(t *testing.T) {
	t.Run("completed work finalizes once", func(t *testing.T) {
		finalized, calls := -1, 0
		n, err := Run(context.Background(), []int{1, 2}, func(context.Context, int) error { return nil }, func(_ context.Context, n int) error { calls++; finalized = n; return nil }, time.Second)
		if n != 2 || finalized != 2 || calls != 1 || err != nil {
			t.Fatalf("accepted=%d finalized=%d calls=%d err=%v", n, finalized, calls, err)
		}
	})

	t.Run("parent cancellation retains cause and finalizes detached work", func(t *testing.T) {
		cause := errors.New("caller stopped")
		ctx, cancel := context.WithCancelCause(context.Background())
		started, release := make(chan struct{}), make(chan struct{})
		done := make(chan struct{})
		var accepted int
		var err error
		finalCalls := 0
		go func() {
			defer close(done)
			accepted, err = Run(ctx, []int{1, 2}, func(_ context.Context, job int) error {
				if job == 1 {
					cancel(cause)
					return nil
				}
				return errors.New("must not start")
			}, func(finalCtx context.Context, n int) error {
				finalCalls++
				if n != 1 || finalCtx.Err() != nil || finalCtx.Done() == nil {
					return errors.New("invalid finalization scope")
				}
				deadline, ok := finalCtx.Deadline()
				if !ok || time.Until(deadline) > time.Second || time.Until(deadline) <= 0 {
					return errors.New("missing bounded deadline")
				}
				close(started)
				<-release
				return nil
			}, 500*time.Millisecond)
		}()
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("finalization did not start")
		}
		select {
		case <-done:
			t.Fatal("Run returned before finalization completed")
		default:
		}
		close(release)
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Run did not join finalization")
		}
		if accepted != 1 || finalCalls != 1 || !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
			t.Fatalf("accepted=%d calls=%d err=%v", accepted, finalCalls, err)
		}
	})

	t.Run("processing and finalization failures remain inspectable", func(t *testing.T) {
		processErr, finalErr := errors.New("apply failed"), errors.New("finalize failed")
		n, err := Run(context.Background(), []int{1, 2}, func(_ context.Context, job int) error {
			if job == 2 {
				return processErr
			}
			return nil
		}, func(context.Context, int) error { return finalErr }, time.Second)
		if n != 1 || !errors.Is(err, processErr) || !errors.Is(err, finalErr) {
			t.Fatalf("accepted=%d err=%v", n, err)
		}
	})

	t.Run("finalization failure alone fails operation", func(t *testing.T) {
		finalErr := errors.New("finalize failed")
		_, err := Run(context.Background(), []int{1}, func(context.Context, int) error { return nil }, func(context.Context, int) error { return finalErr }, time.Second)
		if !errors.Is(err, finalErr) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("empty work does not finalize", func(t *testing.T) {
		calls := 0
		n, err := Run(context.Background(), nil, func(context.Context, int) error { return nil }, func(context.Context, int) error { calls++; return nil }, time.Second)
		if n != 0 || err != nil || calls != 0 {
			t.Fatalf("accepted=%d calls=%d err=%v", n, calls, err)
		}
	})

	t.Run("successful last apply is not vetoed by late cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		n, err := Run(ctx, []int{1}, func(context.Context, int) error { cancel(); return nil }, func(ctx context.Context, n int) error {
			if ctx.Err() != nil || n != 1 {
				return errors.New("finalization failed")
			}
			return nil
		}, time.Second)
		if n != 1 || err != nil {
			t.Fatalf("accepted=%d err=%v", n, err)
		}
	})
}
