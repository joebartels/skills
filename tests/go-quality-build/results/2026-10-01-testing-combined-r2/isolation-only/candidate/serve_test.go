package indexer_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/indexer"
)

var _ func(context.Context, time.Duration, func(context.Context) error, func() error) error = indexer.Serve

func await[T any](t *testing.T, ch <-chan T, operation string) T {
	t.Helper()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case value := <-ch:
		return value
	case <-timer.C:
		t.Fatalf("timed out waiting for %s", operation)
		var zero T
		return zero
	}
}

func TestServeStartupContracts(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Second} {
		t.Run(interval.String(), func(t *testing.T) {
			t.Parallel()
			calls, releases := 0, 0
			err := indexer.Serve(context.Background(), interval, func(context.Context) error { calls++; return nil }, func() error { releases++; return nil })
			if !errors.Is(err, indexer.ErrInvalidInterval) || calls != 0 || releases != 0 {
				t.Fatalf("Serve = %v, calls=%d, releases=%d", err, calls, releases)
			}
		})
	}
	t.Run("already_canceled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		calls, releases := 0, 0
		wantErr := errors.New("release failed")
		err := indexer.Serve(ctx, time.Second, func(context.Context) error { calls++; return nil }, func() error { releases++; return wantErr })
		if !errors.Is(err, wantErr) || calls != 0 || releases != 1 {
			t.Fatalf("Serve = %v, calls=%d, releases=%d", err, calls, releases)
		}
	})
}

func TestServeCallbackAndReleaseErrors(t *testing.T) {
	callbackErr, releaseErr := errors.New("refresh failed"), errors.New("release failed")
	for _, tc := range []struct {
		name     string
		callback error
		release  error
	}{
		{"callback", callbackErr, nil},
		{"both", callbackErr, releaseErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			calls, releases := 0, 0
			err := indexer.Serve(context.Background(), time.Second, func(context.Context) error { calls++; return tc.callback }, func() error { releases++; return tc.release })
			if !errors.Is(err, callbackErr) || (tc.release != nil && !errors.Is(err, releaseErr)) || calls != 1 || releases != 1 {
				t.Fatalf("Serve = %v, calls=%d, releases=%d", err, calls, releases)
			}
		})
	}
}

func TestServeCancellationErrorPrecedence(t *testing.T) {
	independent := errors.New("independent cleanup failure")
	for _, tc := range []struct {
		name   string
		result func(context.Context) error
		want   error
	}{
		{"nil", func(context.Context) error { return nil }, nil},
		{"context_error", func(ctx context.Context) error { return ctx.Err() }, nil},
		{"wrapped_context", func(ctx context.Context) error { return fmt.Errorf("fetch: %w", ctx.Err()) }, nil},
		{"independent_error", func(context.Context) error { return independent }, independent},
		{"joined_independent", func(ctx context.Context) error { return errors.Join(ctx.Err(), independent) }, independent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			releases := 0
			err := indexer.Serve(ctx, time.Hour, func(got context.Context) error {
				if got != ctx {
					return errors.New("callback context changed")
				}
				cancel()
				return tc.result(got)
			}, func() error { releases++; return nil })
			if tc.want == nil && err != nil || tc.want != nil && !errors.Is(err, tc.want) || releases != 1 {
				t.Fatalf("Serve = %v, releases=%d; want %v", err, releases, tc.want)
			}
		})
	}
}

func TestServeCancellationJoinsCleanupBeforeRelease(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	started, cleaning, gate, released := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	finished := make(chan struct{})
	result := make(chan error, 1)
	var gateOnce sync.Once
	openGate := func() { gateOnce.Do(func() { close(gate) }) }
	t.Cleanup(func() {
		cancel()
		openGate()
		await(t, finished, "Serve cleanup join")
	})
	go func() {
		defer close(finished)
		result <- indexer.Serve(ctx, time.Second, func(got context.Context) error {
			if got != ctx {
				return errors.New("callback context changed")
			}
			close(started)
			<-got.Done()
			close(cleaning)
			<-gate
			return got.Err()
		}, func() error { close(released); return nil })
	}()
	await(t, started, "callback startup")
	cancel()
	await(t, cleaning, "callback cooperative cancellation")
	// The callback has observed cancellation and is held in cleanup. This
	// observation window detects release overtaking its completion.
	timer := time.NewTimer(40 * time.Millisecond)
	select {
	case <-released:
		timer.Stop()
		t.Fatal("release overtook callback cleanup")
	case <-finished:
		timer.Stop()
		t.Fatal("Serve returned while cleanup was blocked")
	case <-timer.C:
	}
	openGate()
	if err := await(t, result, "Serve cancellation result"); err != nil {
		t.Fatal(err)
	}
	await(t, released, "release after callback cleanup")
}

func TestServeRecursAfterCompletionWithoutOverlap(t *testing.T) {
	t.Parallel()
	const interval = 50 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	gate := make(chan struct{})
	started := make(chan struct{})
	completed := make(chan time.Time, 1)
	next := make(chan time.Time, 1)
	finished := make(chan struct{})
	result := make(chan error, 1)
	var gateOnce sync.Once
	openGate := func() { gateOnce.Do(func() { close(gate) }) }
	var calls, active, releases atomic.Int32
	stopErr := errors.New("stop after recurrence")
	t.Cleanup(func() { cancel(); openGate(); await(t, finished, "recurrence cleanup join") })
	go func() {
		defer close(finished)
		result <- indexer.Serve(ctx, interval, func(got context.Context) error {
			if got != ctx {
				return errors.New("callback context changed")
			}
			if active.Add(1) != 1 {
				return errors.New("overlapping callbacks")
			}
			defer active.Add(-1)
			if calls.Add(1) == 1 {
				close(started)
				select {
				case <-gate:
				case <-got.Done():
					return got.Err()
				}
				completed <- time.Now()
				return nil
			}
			next <- time.Now()
			return stopErr
		}, func() error { releases.Add(1); return nil })
	}()
	await(t, started, "first callback startup")
	// Hold the first invocation longer than the interval. A ticker whose
	// schedule begins at startup would incorrectly run again immediately.
	timer := time.NewTimer(2 * interval)
	select {
	case <-finished:
		timer.Stop()
		t.Fatal("Serve finished before first callback was released")
	case <-timer.C:
	}
	openGate()
	completion := await(t, completed, "first callback completion")
	second := await(t, next, "second callback startup")
	// Completion is observed just before return; 2ms permits that small
	// measurement gap, while late scheduling is allowed by the contract.
	if elapsed := second.Sub(completion); elapsed < interval-2*time.Millisecond {
		t.Fatalf("next callback began %v after completion; want >= %v", elapsed, interval)
	}
	if err := await(t, result, "recurring Serve result"); !errors.Is(err, stopErr) {
		t.Fatalf("Serve = %v", err)
	}
	await(t, finished, "recurring Serve join")
	if calls.Load() != 2 || active.Load() != 0 || releases.Load() != 1 {
		t.Fatalf("calls=%d, active=%d, releases=%d", calls.Load(), active.Load(), releases.Load())
	}
}

func TestServeIndependentRuns(t *testing.T) {
	t.Parallel()
	type run struct {
		cancel   context.CancelFunc
		starts   chan struct{}
		finished chan struct{}
		result   chan error
		releases atomic.Int32
	}
	start := func() *run {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		r := &run{cancel: cancel, starts: make(chan struct{}, 16), finished: make(chan struct{}), result: make(chan error, 1)}
		t.Cleanup(func() { r.cancel(); await(t, r.finished, "independent Serve cleanup join") })
		go func() {
			defer close(r.finished)
			r.result <- indexer.Serve(ctx, 20*time.Millisecond, func(got context.Context) error {
				if got != ctx {
					return errors.New("callback context changed")
				}
				select {
				case r.starts <- struct{}{}:
				default:
				}
				return nil
			}, func() error { r.releases.Add(1); return nil })
		}()
		return r
	}
	first, second := start(), start()
	await(t, first.starts, "first instance startup")
	await(t, second.starts, "second instance startup")
	first.cancel()
	if err := await(t, first.result, "first instance stop"); err != nil {
		t.Fatal(err)
	}
	await(t, first.finished, "first instance join")
	// Drain prior observations, then observe a fresh callback after the first
	// run is fully joined. The second context remains independently active.
	for len(second.starts) > 0 {
		<-second.starts
	}
	await(t, second.starts, "second instance recurrence after first stops")
	second.cancel()
	if err := await(t, second.result, "second instance stop"); err != nil {
		t.Fatal(err)
	}
	await(t, second.finished, "second instance join")
	if first.releases.Load() != 1 || second.releases.Load() != 1 {
		t.Fatalf("release counts = %d, %d", first.releases.Load(), second.releases.Load())
	}
}

func TestServeParentCancellationCause(t *testing.T) {
	t.Parallel()
	cause := errors.New("parent shutdown")
	independent := errors.New("independent failure")
	for _, tc := range []struct {
		name             string
		joinIndependent  bool
		joinContextError bool
	}{
		{"ordinary_cause", false, false},
		{"cause_with_context_error", false, true},
		{"cause_with_independent_error", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancelCause(context.Background())
			defer cancel(nil)
			releases := 0
			err := indexer.Serve(ctx, time.Second, func(got context.Context) error {
				cancel(cause)
				result := fmt.Errorf("fetch: %w", context.Cause(got))
				if tc.joinContextError {
					return errors.Join(result, got.Err())
				}
				if tc.joinIndependent {
					return errors.Join(result, independent)
				}
				return result
			}, func() error { releases++; return nil })
			if !tc.joinIndependent && err != nil || tc.joinIndependent && !errors.Is(err, independent) || releases != 1 {
				t.Fatalf("Serve = %v, releases=%d", err, releases)
			}
		})
	}
}
