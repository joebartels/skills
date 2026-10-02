package finalize

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunSequentialAcceptanceAndFinalization(t *testing.T) {
	var applied []int
	finalized := 0
	n, err := Run(context.Background(), []int{3, 5}, func(_ context.Context, job int) error {
		applied = append(applied, job)
		return nil
	}, func(_ context.Context, count int) error {
		finalized = count
		return nil
	}, time.Second)
	if err != nil || n != 2 || finalized != 2 || len(applied) != 2 || applied[0] != 3 || applied[1] != 5 {
		t.Fatalf("accepted=%d finalized=%d applied=%v err=%v", n, finalized, applied, err)
	}
}

func TestRunStopsOnApplyFailureAndJoinsFinalizationFailure(t *testing.T) {
	workErr := errors.New("apply failed")
	finishErr := errors.New("finalize failed")
	var applied []int
	finalized := 0
	n, err := Run(context.Background(), []int{1, 2, 3}, func(_ context.Context, job int) error {
		applied = append(applied, job)
		if job == 2 {
			return workErr
		}
		return nil
	}, func(_ context.Context, count int) error { finalized = count; return finishErr }, time.Second)
	if n != 1 || finalized != 1 || len(applied) != 2 || applied[1] != 2 {
		t.Fatalf("accepted=%d finalized=%d applied=%v", n, finalized, applied)
	}
	if !errors.Is(err, workErr) || !errors.Is(err, finishErr) {
		t.Fatalf("err=%v, want both work and finalization causes", err)
	}
}

func TestRunPreCanceledContextDoesNoWork(t *testing.T) {
	base, cancel := context.WithCancelCause(context.Background())
	cause := errors.New("shutdown requested")
	cancel(cause)
	called := false
	n, err := Run(base, []int{1}, func(context.Context, int) error { called = true; return nil }, func(context.Context, int) error { called = true; return nil }, time.Second)
	if n != 0 || called || !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
		t.Fatalf("accepted=%d called=%v err=%v", n, called, err)
	}
}

func TestRunCancellationBetweenJobsRetainsAcceptedPrefixAndCause(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	cause := errors.New("stop remaining jobs")
	applied := 0
	finalized := 0
	n, err := Run(ctx, []int{1, 2}, func(context.Context, int) error {
		applied++
		cancel(cause)
		return nil
	}, func(_ context.Context, count int) error { finalized = count; return nil }, time.Second)
	if n != 1 || applied != 1 || finalized != 1 || !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
		t.Fatalf("accepted=%d applied=%d finalized=%d err=%v", n, applied, finalized, err)
	}
}

func TestRunSuccessfulLastApplyIsNotVetoedByLateCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finalized := false
	n, err := Run(ctx, []int{1}, func(context.Context, int) error { cancel(); return nil }, func(finalCtx context.Context, count int) error {
		finalized = count == 1 && finalCtx.Err() == nil
		return nil
	}, time.Second)
	if n != 1 || !finalized || err != nil {
		t.Fatalf("accepted=%d finalized=%v err=%v", n, finalized, err)
	}
}

func TestRunFinalizationSurvivesParentCancellationAndJoinsBeforeReturn(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	entered := make(chan struct{})
	completed := false
	started := time.Now()
	n, err := Run(parent, []int{1}, func(context.Context, int) error { cancelParent(); return nil }, func(ctx context.Context, count int) error {
		if count != 1 || ctx.Err() != nil {
			t.Errorf("finalization count=%d ctx err=%v", count, ctx.Err())
		}
		close(entered)
		<-ctx.Done()
		completed = true
		return ctx.Err()
	}, 30*time.Millisecond)
	select {
	case <-entered:
	default:
		t.Fatal("finalizer was not called")
	}
	if !completed || n != 1 || !errors.Is(err, context.DeadlineExceeded) || time.Since(started) < 20*time.Millisecond {
		t.Fatalf("accepted=%d completed=%v elapsed=%v err=%v", n, completed, time.Since(started), err)
	}
}

func TestRunFinalizationFailureFailsSuccessfulProcessing(t *testing.T) {
	finishErr := errors.New("receipt write failed")
	n, err := Run(context.Background(), []int{1}, func(context.Context, int) error { return nil }, func(context.Context, int) error { return finishErr }, time.Second)
	if n != 1 || !errors.Is(err, finishErr) {
		t.Fatalf("accepted=%d err=%v", n, err)
	}
}

func TestRunDoesNotFinalizeWithoutAcceptedWork(t *testing.T) {
	called := false
	workErr := errors.New("no accepted work")
	n, err := Run(context.Background(), []int{1}, func(context.Context, int) error { return workErr }, func(context.Context, int) error { called = true; return nil }, time.Second)
	if n != 0 || called || !errors.Is(err, workErr) {
		t.Fatalf("accepted=%d finalize called=%v err=%v", n, called, err)
	}
}
