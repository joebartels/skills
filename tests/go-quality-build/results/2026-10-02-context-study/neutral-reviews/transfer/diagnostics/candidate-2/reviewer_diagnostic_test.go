package finalize

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Reviewer diagnostic only: not part of either authored test suite or efficacy score.
func TestDiagnosticIndependentApplyFailureWithCancellation(t *testing.T) {
	parentCause := errors.New("caller stopped")
	workFailure := errors.New("accepted job persistence failed")
	finalFailure := errors.New("receipt persistence failed")
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	finalCalls := 0
	n, err := Run(ctx, []int{1, 2, 3}, func(_ context.Context, job int) error {
		if job == 2 {
			cancel(parentCause)
			return errors.Join(workFailure, context.Canceled)
		}
		return nil
	}, func(ctx context.Context, count int) error {
		finalCalls++
		if count != 1 || ctx.Err() != nil {
			t.Errorf("finalization count=%d err=%v", count, ctx.Err())
		}
		return finalFailure
	}, time.Second)
	if n != 1 || finalCalls != 1 || !errors.Is(err, context.Canceled) || !errors.Is(err, parentCause) || !errors.Is(err, workFailure) || !errors.Is(err, finalFailure) {
		t.Fatalf("accepted=%d finalCalls=%d canceled=%v parentCause=%v workFailure=%v finalFailure=%v err=%v", n, finalCalls, errors.Is(err, context.Canceled), errors.Is(err, parentCause), errors.Is(err, workFailure), errors.Is(err, finalFailure), err)
	}
}

func TestDiagnosticFinalizationScopeReleased(t *testing.T) {
	var captured context.Context
	n, err := Run(context.Background(), []int{1}, func(context.Context, int) error { return nil }, func(ctx context.Context, count int) error {
		captured = ctx
		if _, ok := ctx.Deadline(); !ok {
			t.Error("finalization has no deadline")
		}
		return nil
	}, time.Hour)
	if n != 1 || err != nil || captured == nil || captured.Err() != context.Canceled {
		t.Fatalf("accepted=%d err=%v scope=%v", n, err, captured)
	}
}
