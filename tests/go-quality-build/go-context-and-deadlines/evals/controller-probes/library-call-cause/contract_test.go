package process

import (
	"context"
	"errors"
	"testing"
	"time"
)

type listCause []string

func (e listCause) Error() string { return "custom cancellation" }

func TestCanceledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(listCause{"stop", "request"})
	for _, jobs := range [][]int{nil, {1, 2}} {
		calls := 0
		n, err := Process(ctx, jobs, func(context.Context, int) error { calls++; return nil })
		var cause listCause
		if calls != 0 || n != 0 || !errors.Is(err, context.Canceled) || !errors.As(err, &cause) || len(cause) != 2 || cause[0] != "stop" || cause[1] != "request" {
			t.Fatalf("calls=%d accepted=%d err=%v cause=%v", calls, n, err, cause)
		}
	}
}

func TestCustomCauseAndIndependentFailure(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	entered, release := make(chan struct{}), make(chan struct{})
	type outcome struct {
		n   int
		err error
	}
	done := make(chan outcome, 1)
	failure := errors.New("independent apply failure")
	go func() {
		n, err := Process(ctx, []int{1, 2, 3}, func(_ context.Context, job int) error {
			if job == 2 {
				close(entered)
				<-release
				return failure
			}
			return nil
		})
		done <- outcome{n, err}
	}()
	select {
	case <-entered:
	case got := <-done:
		t.Fatalf("returned before held callback: %+v", got)
	case <-time.After(2 * time.Second):
		cancel(nil)
		close(release)
		t.Fatal("callback did not start")
	}
	cancel(listCause{"owned", "cause"})
	close(release)
	select {
	case got := <-done:
		var cause listCause
		if got.n != 1 || !errors.Is(got.err, failure) || !errors.Is(got.err, context.Canceled) || !errors.As(got.err, &cause) || len(cause) != 2 || cause[0] != "owned" || cause[1] != "cause" {
			t.Fatalf("accepted=%d err=%v cause=%v", got.n, got.err, cause)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("callback returned but Process did not")
	}
}

func TestResultDecisionOrder(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	cause := errors.New("between jobs")
	calls := 0
	n, err := Process(ctx, []int{1, 2}, func(context.Context, int) error { calls++; cancel(cause); return nil })
	if n != 1 || calls != 1 || !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
		t.Fatalf("accepted=%d calls=%d err=%v", n, calls, err)
	}
	ctx, cancel = context.WithCancelCause(context.Background())
	n, err = Process(ctx, []int{1}, func(context.Context, int) error { cancel(cause); return nil })
	if n != 1 || err != nil {
		t.Fatalf("completed success invalidated: accepted=%d err=%v", n, err)
	}
	n, err = Process(context.Background(), nil, func(context.Context, int) error { t.Fatal("empty input invoked callback"); return nil })
	if n != 0 || err != nil {
		t.Fatalf("empty: accepted=%d err=%v", n, err)
	}
}

func TestCauseReturnedByCallback(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Fatalf("Process panicked for a legal non-comparable cause: %v", p)
		}
	}()
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	n, err := Process(ctx, []int{1}, func(context.Context, int) error {
		cancel(listCause{"callback", "cause"})
		return context.Cause(ctx)
	})
	var cause listCause
	if n != 0 || !errors.Is(err, context.Canceled) || !errors.As(err, &cause) || len(cause) != 2 || cause[0] != "callback" || cause[1] != "cause" {
		t.Fatalf("accepted=%d err=%v cause=%v", n, err, cause)
	}
}
