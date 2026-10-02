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

const lifecycleDeadline = 2 * time.Second

type serveResult struct {
	done chan struct{}
	err  error
}

func launchServe(ctx context.Context, interval time.Duration, refresh func(context.Context) error, release func() error) *serveResult {
	result := &serveResult{done: make(chan struct{})}
	go func() {
		result.err = indexer.Serve(ctx, interval, refresh, release)
		close(result.done)
	}()
	return result
}

func await(t *testing.T, event <-chan struct{}, label string) {
	t.Helper()
	select {
	case <-event:
	case <-time.After(lifecycleDeadline):
		t.Fatalf("timed out waiting for %s", label)
	}
}

func cleanupRun(t *testing.T, cancel context.CancelFunc, unblock func(), result *serveResult) {
	t.Helper()
	t.Cleanup(func() {
		cancel()
		if unblock != nil {
			unblock()
		}
		select {
		case <-result.done:
		case <-time.After(lifecycleDeadline):
			t.Error("Serve did not join during cleanup")
		}
	})
}

func TestServeInvalidIntervalsNeverInvokeCallbacks(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Second} {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		calls, releases := 0, 0
		err := indexer.Serve(ctx, interval, func(context.Context) error {
			calls++
			return nil
		}, func() error {
			releases++
			return nil
		})
		if !errors.Is(err, indexer.ErrInvalidInterval) || calls != 0 || releases != 0 {
			t.Fatalf("interval %s: result = %v, callbacks = %d, releases = %d", interval, err, calls, releases)
		}
	}
}

func TestServeAlreadyCanceledReleasesWithoutStarting(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	failure := errors.New("release failed")
	calls, releases := 0, 0
	err := indexer.Serve(ctx, time.Hour, func(context.Context) error { calls++; return nil }, func() error { releases++; return failure })
	if !errors.Is(err, failure) || calls != 0 || releases != 1 {
		t.Fatalf("result = %v, callbacks = %d, releases = %d", err, calls, releases)
	}
}

func TestServeRecursAfterCompletionWithoutOverlap(t *testing.T) {
	const interval = 40 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	firstStarted := make(chan struct{})
	finishFirst := make(chan struct{})
	secondObserved := make(chan struct{})
	var finishOnce, observedOnce sync.Once
	finish := func() { finishOnce.Do(func() { close(finishFirst) }) }
	observe := func() { observedOnce.Do(func() { close(secondObserved) }) }
	unblock := func() { finish(); observe() }
	secondStarted := make(chan time.Time, 1)
	var calls, active atomic.Int32
	var overlap, wrongContext atomic.Bool
	releases := 0
	result := launchServe(ctx, interval, func(got context.Context) error {
		if got != ctx {
			wrongContext.Store(true)
		}
		if active.Add(1) != 1 {
			overlap.Store(true)
		}
		defer active.Add(-1)
		switch calls.Add(1) {
		case 1:
			close(firstStarted)
			select {
			case <-finishFirst:
			case <-got.Done():
			}
		case 2:
			secondStarted <- time.Now()
			<-secondObserved
			cancel()
		default:
			return errors.New("unexpected extra callback")
		}
		return nil
	}, func() error { releases++; return nil })
	cleanupRun(t, cancel, unblock, result)
	await(t, firstStarted, "immediate first callback")
	// Hold work beyond the interval: a ticker measured from its start is wrong.
	select {
	case <-secondStarted:
		t.Fatal("callback overlapped the held first invocation")
	case <-result.done:
		t.Fatalf("Serve exited early: %v", result.err)
	case <-time.After(2 * interval):
	}
	completed := time.Now()
	finish()
	select {
	case started := <-secondStarted:
		if elapsed := started.Sub(completed); elapsed < interval {
			t.Errorf("second callback began %s after first completion; want at least %s", elapsed, interval)
		}
	case <-result.done:
		t.Fatalf("Serve exited before recurrence: %v", result.err)
	case <-time.After(lifecycleDeadline):
		t.Fatal("no recurring callback")
	}
	observe()
	await(t, result.done, "canceled recurring run")
	if result.err != nil || calls.Load() != 2 || releases != 1 || overlap.Load() || wrongContext.Load() {
		t.Fatalf("result = %v, calls = %d, releases = %d, overlap = %v, wrong context = %v", result.err, calls.Load(), releases, overlap.Load(), wrongContext.Load())
	}
}

func TestServeParentCancellationCauseAndDeadline(t *testing.T) {
	t.Run("cancellation cause", func(t *testing.T) {
		ctx, cancel := context.WithCancelCause(context.Background())
		started := make(chan struct{})
		releases := 0
		result := launchServe(ctx, time.Hour, func(got context.Context) error {
			close(started)
			<-got.Done()
			return fmt.Errorf("request stopped: %w", errors.Join(got.Err(), context.Cause(got)))
		}, func() error { releases++; return nil })
		cleanupRun(t, func() { cancel(nil) }, nil, result)
		await(t, started, "callback start")
		cancel(errors.New("parent requested shutdown"))
		await(t, result.done, "parent cancellation with a cause")
		if result.err != nil || releases != 1 {
			t.Fatalf("result = %v, releases = %d", result.err, releases)
		}
	})
	t.Run("deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		releases := 0
		result := launchServe(ctx, time.Hour, func(got context.Context) error {
			<-got.Done()
			return got.Err()
		}, func() error { releases++; return nil })
		cleanupRun(t, cancel, nil, result)
		await(t, result.done, "parent deadline")
		if result.err != nil || releases != 1 {
			t.Fatalf("result = %v, releases = %d", result.err, releases)
		}
	})
}

func TestServeWaitsForCanceledCleanupAndPreservesErrors(t *testing.T) {
	callbackFailure := errors.New("callback failed")
	releaseFailure := errors.New("release failed")
	cases := []struct {
		name       string
		callback   func(context.Context) error
		releaseErr error
		want       []error
	}{
		{"nil callback", func(context.Context) error { return nil }, nil, nil},
		{"context callback", func(ctx context.Context) error { return ctx.Err() }, nil, nil},
		{"wrapped context callback", func(ctx context.Context) error { return fmt.Errorf("fetch: %w", ctx.Err()) }, nil, nil},
		{"independent callback", func(context.Context) error { return callbackFailure }, nil, []error{callbackFailure}},
		{"release failure", func(context.Context) error { return nil }, releaseFailure, []error{releaseFailure}},
		{"callback and release", func(context.Context) error { return callbackFailure }, releaseFailure, []error{callbackFailure, releaseFailure}},
		{"joined callback causes", func(ctx context.Context) error { return errors.Join(ctx.Err(), callbackFailure) }, releaseFailure, []error{context.Canceled, callbackFailure, releaseFailure}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			started, cleaning, cleaned := make(chan struct{}), make(chan struct{}), make(chan struct{})
			finish := make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(finish) }) }
			released := make(chan struct{})
			var overtook atomic.Bool
			calls, releases := 0, 0
			result := launchServe(ctx, time.Hour, func(got context.Context) error {
				calls++
				if got != ctx {
					return errors.New("caller context not forwarded")
				}
				close(started)
				<-got.Done()
				close(cleaning)
				<-finish
				close(cleaned)
				return tc.callback(got)
			}, func() error {
				releases++
				select {
				case <-cleaned:
				default:
					overtook.Store(true)
				}
				close(released)
				return tc.releaseErr
			})
			cleanupRun(t, cancel, unblock, result)
			await(t, started, "callback start")
			cancel()
			await(t, cleaning, "cooperative cancellation cleanup")
			select {
			case <-released:
				t.Fatal("release overtook callback cleanup")
			case <-result.done:
				t.Fatal("Serve returned before callback cleanup")
			case <-time.After(30 * time.Millisecond):
			}
			unblock()
			await(t, result.done, "cleanup completion and join")
			if len(tc.want) == 0 && result.err != nil {
				t.Fatalf("ordinary cancellation = %v; want nil", result.err)
			}
			for _, cause := range tc.want {
				if !errors.Is(result.err, cause) {
					t.Errorf("result = %v; lost cause %v", result.err, cause)
				}
			}
			if calls != 1 || releases != 1 || overtook.Load() {
				t.Fatalf("calls = %d, releases = %d, release overtook = %v", calls, releases, overtook.Load())
			}
		})
	}
}

func TestServeImmediateFailureReleasesOnce(t *testing.T) {
	callbackFailure, releaseFailure := errors.New("refresh failed"), errors.New("release failed")
	calls, releases := 0, 0
	err := indexer.Serve(context.Background(), time.Hour, func(context.Context) error { calls++; return callbackFailure }, func() error { releases++; return releaseFailure })
	if !errors.Is(err, callbackFailure) || !errors.Is(err, releaseFailure) || calls != 1 || releases != 1 {
		t.Fatalf("result = %v, calls = %d, releases = %d", err, calls, releases)
	}
}

func TestServeInvocationsAreIndependentlyControlled(t *testing.T) {
	ctxA, cancelA := context.WithCancel(context.Background())
	ctxB, cancelB := context.WithCancel(context.Background())
	aCalls, bCalls := make(chan struct{}, 16), make(chan struct{}, 16)
	callback := func(events chan<- struct{}) func(context.Context) error {
		return func(context.Context) error {
			select {
			case events <- struct{}{}:
			default:
			}
			return nil
		}
	}
	a := launchServe(ctxA, 10*time.Millisecond, callback(aCalls), func() error { return nil })
	b := launchServe(ctxB, 10*time.Millisecond, callback(bCalls), func() error { return nil })
	cleanupRun(t, cancelA, nil, a)
	cleanupRun(t, cancelB, nil, b)
	await(t, aCalls, "first run startup")
	await(t, bCalls, "second run startup")
	cancelA()
	await(t, a.done, "first run stop")
	for len(bCalls) != 0 {
		<-bCalls
	}
	select {
	case <-bCalls:
	case <-b.done:
		t.Fatalf("second run stopped with first: %v", b.err)
	case <-time.After(lifecycleDeadline):
		t.Fatal("second run did not continue after first stopped")
	}
	cancelB()
	await(t, b.done, "second run stop")
	if a.err != nil || b.err != nil {
		t.Fatalf("run errors: %v, %v", a.err, b.err)
	}
}
