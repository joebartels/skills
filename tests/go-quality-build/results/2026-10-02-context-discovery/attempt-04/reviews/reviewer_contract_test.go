package process

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestReviewerCanceledEmpty(t *testing.T) {
	for _, jobs := range [][]int{nil, {}} {
		cause := sliceCause{values: []int{21}}
		ctx, cancel := context.WithCancelCause(context.Background())
		cancel(cause)
		calls := 0
		n, err := Process(ctx, jobs, func(context.Context, int) error { calls++; return nil })
		var got sliceCause
		if n != 0 || calls != 0 || !errors.Is(err, context.Canceled) || !errors.As(err, &got) || !reflect.DeepEqual(got, cause) {
			t.Errorf("jobs=%#v accepted=%d calls=%d err=%v cause=%#v", jobs, n, calls, err, got)
		}
	}
}

func TestReviewerCauseReturnedByCallback(t *testing.T) {
	cause := sliceCause{values: []int{22}}
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	calls := 0
	n, err := Process(ctx, []int{1, 2, 3}, func(context.Context, int) error {
		calls++
		if calls == 1 { return nil }
		cancel(cause)
		return cause
	})
	var got sliceCause
	if n != 1 || calls != 2 || !errors.Is(err, context.Canceled) || !errors.As(err, &got) || !reflect.DeepEqual(got, cause) {
		t.Fatalf("accepted=%d calls=%d err=%v cause=%#v", n, calls, err, got)
	}
}

func TestReviewerDeadlineAndIndependentFailure(t *testing.T) {
	cause := errors.New("deadline cause")
	ctx, cancel := context.WithDeadlineCause(context.Background(), time.Unix(0, 0), cause)
	defer cancel()
	n, err := Process(ctx, []int{1}, func(context.Context, int) error { t.Fatal("callback called"); return nil })
	if n != 0 || !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, cause) {
		t.Fatalf("accepted=%d err=%v", n, err)
	}
	failure := errors.New("independent failure")
	ctx2, cancel2 := context.WithCancelCause(context.Background())
	defer cancel2(nil)
	calls := 0
	n, err = Process(ctx2, []int{1, 2, 3}, func(context.Context, int) error {
		calls++
		if calls == 1 { return nil }
		cancel2(cause)
		return failure
	})
	if n != 1 || calls != 2 || !errors.Is(err, failure) || !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
		t.Fatalf("accepted=%d calls=%d err=%v", n, calls, err)
	}
}

func TestReviewerSuccessThenCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var seen []int
	n, err := Process(ctx, []int{1, 2, 3}, func(_ context.Context, job int) error { seen = append(seen, job); return nil })
	cancel()
	if n != 3 || err != nil || !reflect.DeepEqual(seen, []int{1, 2, 3}) {
		t.Fatalf("accepted=%d err=%v seen=%v", n, err, seen)
	}
}

func TestReviewerConcurrentCooperativeCancellation(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	cause := errors.New("concurrent cancellation cause")
	failure := errors.New("cooperative callback failure")
	entered := make(chan struct{})
	type result struct { accepted int; err error }
	completed := make(chan result, 1)
	go func() {
		n, err := Process(ctx, []int{1, 2}, func(ctx context.Context, _ int) error {
			close(entered)
			<-ctx.Done()
			return failure
		})
		completed <- result{n, err}
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second): t.Fatal("callback did not start")
	}
	cancel(cause)
	select {
	case got := <-completed:
		if got.accepted != 0 || !errors.Is(got.err, failure) || !errors.Is(got.err, context.Canceled) || !errors.Is(got.err, cause) {
			t.Fatalf("accepted=%d err=%v", got.accepted, got.err)
		}
	case <-time.After(3 * time.Second): t.Fatal("Process did not return cooperatively")
	}
}
