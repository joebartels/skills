package receipt

import (
	"context"
	"errors"
	"testing"
	"time"
)

type recoveryKey struct{}
type recoveryCause struct{ labels []string }

func (recoveryCause) Error() string { return "caller stopped" }

func TestRecoveryReceiptCancellation(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.WithValue(context.Background(), recoveryKey{}, "import-7"))
	defer cancel(nil)
	finishErr := errors.New("receipt fault")
	applyCalls, finishCalls := 0, 0
	n, err := ApplyBatch(ctx, []string{"one", "two"}, time.Second,
		func(context.Context, string) error {
			applyCalls++
			cancel(recoveryCause{[]string{"shutdown"}})
			return nil
		}, func(final context.Context, count int) error {
			finishCalls++
			deadline, ok := final.Deadline()
			if final.Err() != nil || !ok || deadline.After(time.Now().Add(time.Second)) || final.Value(recoveryKey{}) != "import-7" || count != 1 {
				t.Error("receipt lost its live bounded metadata scope or accepted count")
			}
			return errors.Join(context.Canceled, finishErr)
		})
	var cause recoveryCause
	if n != 1 || applyCalls != 1 || finishCalls != 1 || !errors.Is(err, context.Canceled) || !errors.Is(err, finishErr) || !errors.As(err, &cause) {
		t.Fatalf("n=%d apply=%d finish=%d err=%v", n, applyCalls, finishCalls, err)
	}
}

func TestRecoveryReceiptIndependentAndSuccess(t *testing.T) {
	applyErr, finishErr := errors.New("apply fault"), errors.New("receipt fault")
	finished := 0
	n, err := ApplyBatch(context.Background(), []string{"one", "two"}, time.Second,
		func(_ context.Context, item string) error {
			if item == "two" {
				return errors.Join(context.DeadlineExceeded, applyErr)
			}
			return nil
		}, func(context.Context, int) error { finished++; return finishErr })
	if n != 1 || finished != 1 || !errors.Is(err, applyErr) || !errors.Is(err, finishErr) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("n=%d finished=%d err=%v", n, finished, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	n, err = ApplyBatch(ctx, []string{"one"}, time.Second,
		func(context.Context, string) error { cancel(); return nil },
		func(final context.Context, _ int) error {
			if final.Err() != nil {
				t.Error("completed accepted item could not finalize")
			}
			return nil
		})
	if n != 1 || err != nil {
		t.Fatalf("completed n=%d err=%v", n, err)
	}
	bad := func(context.Context, string) error { t.Error("empty apply called"); return nil }
	n, err = ApplyBatch(ctx, nil, time.Second, bad, func(context.Context, int) error { t.Error("empty receipt called"); return nil })
	if n != 0 || err != nil {
		t.Fatalf("empty n=%d err=%v", n, err)
	}
}
