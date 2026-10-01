package sweeper_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/sweeper"
)

// Check the public function type as well as ordinary calls.
var _ func(context.Context, time.Duration, func(context.Context) error, func() error) error = sweeper.Run

const waitLimit = 2 * time.Second

type runningHost struct {
	done chan struct{}
	err  error
}

func startHost(t *testing.T, ctx context.Context, cancel context.CancelFunc, interval time.Duration, sweep func(context.Context) error, release func() error) *runningHost {
	t.Helper()
	host := &runningHost{done: make(chan struct{})}
	go func() {
		host.err = sweeper.Run(ctx, interval, sweep, release)
		close(host.done)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-host.done:
		case <-time.After(waitLimit):
			t.Error("host did not finish during test cleanup")
		}
	})
	return host
}

func waitFor[T any](t *testing.T, description string, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(waitLimit):
		t.Fatalf("timed out waiting for %s", description)
		var zero T
		return zero
	}
}

func waitResult(t *testing.T, host *runningHost) error {
	t.Helper()
	waitFor(t, "Run to return", host.done)
	return host.err
}

func TestRunImmediateThenFullIntervalAfterCompletion(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	const interval = 40 * time.Millisecond
	starts := make(chan time.Time, 8)
	firstCompleted := make(chan time.Time, 1)
	finishFirst := make(chan struct{})
	var calls, active, releases atomic.Int32
	var overlap atomic.Bool
	host := startHost(t, ctx, cancel, interval, func(ctx context.Context) error {
		if active.Add(1) != 1 {
			overlap.Store(true)
		}
		defer active.Add(-1)
		call := calls.Add(1)
		select {
		case starts <- time.Now():
		case <-ctx.Done():
			return nil
		}
		if call == 1 {
			select {
			case <-finishFirst:
				firstCompleted <- time.Now()
			case <-ctx.Done():
			}
			return nil
		}
		cancel()
		return nil
	}, func() error { releases.Add(1); return nil })

	waitFor(t, "immediate first sweep", starts)
	// Hold the callback past the interval to expose accumulated ticker ticks.
	select {
	case <-starts:
		t.Fatal("another sweep started while the first callback was blocked")
	case <-host.done:
		t.Fatalf("Run returned while the first callback was blocked: %v", host.err)
	case <-time.After(2 * interval):
	}
	close(finishFirst)
	completed := waitFor(t, "first callback completion", firstCompleted)
	secondStarted := waitFor(t, "second sweep", starts)
	if elapsed := secondStarted.Sub(completed); elapsed < interval {
		t.Errorf("next sweep started after %s; want at least %s after completion", elapsed, interval)
	}
	if err := waitResult(t, host); err != nil {
		t.Fatalf("Run = %v; want nil after cancellation", err)
	}
	if overlap.Load() || calls.Load() != 2 || releases.Load() != 1 {
		t.Fatalf("overlap = %v, calls = %d, releases = %d; want false, 2, 1", overlap.Load(), calls.Load(), releases.Load())
	}
}

func TestRunStartsImmediatelyWithLongInterval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{}, 1)
	host := startHost(t, ctx, cancel, time.Hour, func(ctx context.Context) error {
		started <- struct{}{}
		<-ctx.Done()
		return nil
	}, func() error { return nil })
	waitFor(t, "first sweep without waiting an hour", started)
	cancel()
	if err := waitResult(t, host); err != nil {
		t.Fatalf("Run = %v; want nil", err)
	}
}

func TestRunCancellationWaitsForCallbackCleanupBeforeRelease(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{}, 1)
	cleaning := make(chan struct{})
	finishCleanup := make(chan struct{})
	released := make(chan struct{}, 1)
	var cleaned, releaseSawCleanup, releaseSawCancellation atomic.Bool
	var releases atomic.Int32
	var child context.Context
	host := startHost(t, ctx, cancel, time.Hour, func(ctx context.Context) error {
		child = ctx
		started <- struct{}{}
		<-ctx.Done()
		close(cleaning)
		<-finishCleanup
		cleaned.Store(true)
		return nil
	}, func() error {
		releases.Add(1)
		releaseSawCleanup.Store(cleaned.Load())
		releaseSawCancellation.Store(child != nil && child.Err() != nil)
		released <- struct{}{}
		return nil
	})
	var finishOnce sync.Once
	finish := func() { finishOnce.Do(func() { close(finishCleanup) }) }
	// Open the cleanup gate before startHost's cancel-and-join cleanup on failure.
	t.Cleanup(finish)
	waitFor(t, "first sweep", started)
	cancel()
	waitFor(t, "callback entering cancellation cleanup", cleaning)
	select {
	case <-released:
		t.Fatal("release ran while callback cleanup was blocked")
	case <-host.done:
		t.Fatalf("Run returned while callback cleanup was blocked: %v", host.err)
	case <-time.After(30 * time.Millisecond):
	}
	finish()
	if err := waitResult(t, host); err != nil {
		t.Fatalf("Run = %v; want nil", err)
	}
	if releases.Load() != 1 || !releaseSawCleanup.Load() || !releaseSawCancellation.Load() {
		t.Fatalf("releases = %d, saw cleanup = %v, saw cancellation = %v; want 1, true, true", releases.Load(), releaseSawCleanup.Load(), releaseSawCancellation.Load())
	}
}

func TestRunAlreadyCanceledSkipsSweepAndReleasesOnce(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var calls, releases atomic.Int32
	host := startHost(t, ctx, cancel, time.Second, func(context.Context) error {
		calls.Add(1)
		return nil
	}, func() error { releases.Add(1); return nil })
	if err := waitResult(t, host); err != nil || calls.Load() != 0 || releases.Load() != 1 {
		t.Fatalf("Run = %v, calls = %d, releases = %d; want nil, 0, 1", err, calls.Load(), releases.Load())
	}
}

func TestRunCancellationBetweenCycles(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	completed := make(chan struct{}, 1)
	var calls, releases atomic.Int32
	host := startHost(t, ctx, cancel, time.Hour, func(context.Context) error {
		calls.Add(1)
		completed <- struct{}{}
		return nil
	}, func() error { releases.Add(1); return nil })
	waitFor(t, "successful callback", completed)
	cancel()
	if err := waitResult(t, host); err != nil || calls.Load() != 1 || releases.Load() != 1 {
		t.Fatalf("Run = %v, calls = %d, releases = %d; want nil, 1, 1", err, calls.Load(), releases.Load())
	}
}

type callbackError struct{}

func (*callbackError) Error() string { return "callback failed" }

func TestRunFailureStopsAndJoinsReleaseError(t *testing.T) {
	for _, tc := range []struct {
		name   string
		failAt int32
	}{
		{"immediate", 1}, {"after_successful_cycles", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			wantSweep := &callbackError{}
			wantRelease := errors.New("release failed")
			var calls, releases atomic.Int32
			var cleaned, releaseSawCleanup, releaseSawCancellation atomic.Bool
			var child context.Context
			host := startHost(t, ctx, cancel, time.Millisecond, func(ctx context.Context) error {
				child = ctx
				if calls.Add(1) == tc.failAt {
					defer cleaned.Store(true)
					return wantSweep
				}
				return nil
			}, func() error {
				releases.Add(1)
				releaseSawCleanup.Store(cleaned.Load())
				releaseSawCancellation.Store(child != nil && child.Err() != nil)
				return wantRelease
			})
			err := waitResult(t, host)
			if !errors.Is(err, wantSweep) || !errors.Is(err, wantRelease) {
				t.Fatalf("Run = %v; want both callback and release errors", err)
			}
			var gotSweep *callbackError
			if !errors.As(err, &gotSweep) || gotSweep != wantSweep {
				t.Fatalf("callback error type/identity not preserved: %v", err)
			}
			if calls.Load() != tc.failAt || releases.Load() != 1 || !releaseSawCleanup.Load() || !releaseSawCancellation.Load() {
				t.Fatalf("calls = %d, releases = %d, saw cleanup = %v, saw cancellation = %v; want %d, 1, true, true", calls.Load(), releases.Load(), releaseSawCleanup.Load(), releaseSawCancellation.Load(), tc.failAt)
			}
		})
	}
}

func TestRunCanceledCallbackResult(t *testing.T) {
	independent := errors.New("independent callback failure")
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "normal_cancellation"},
		{name: "context_error", err: context.Canceled},
		{name: "independent_error", err: independent},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			started := make(chan struct{}, 1)
			wantRelease := errors.New("release failed")
			var calls, releases atomic.Int32
			host := startHost(t, ctx, cancel, time.Hour, func(ctx context.Context) error {
				calls.Add(1)
				started <- struct{}{}
				<-ctx.Done()
				return tc.err
			}, func() error { releases.Add(1); return wantRelease })
			waitFor(t, "callback start", started)
			cancel()
			err := waitResult(t, host)
			if !errors.Is(err, wantRelease) || (tc.err != nil && !errors.Is(err, tc.err)) {
				t.Fatalf("Run = %v; want callback result %v joined with release error", err, tc.err)
			}
			if tc.err == nil && errors.Is(err, context.Canceled) {
				t.Fatalf("normal cancellation contributed an error: %v", err)
			}
			if calls.Load() != 1 || releases.Load() != 1 {
				t.Fatalf("calls = %d, releases = %d; want 1, 1", calls.Load(), releases.Load())
			}
		})
	}
}

func TestRunRejectsNonpositiveIntervalsWithoutCallbacks(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Nanosecond, -time.Hour} {
		t.Run(interval.String(), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			var calls, releases atomic.Int32
			host := startHost(t, ctx, cancel, interval, func(context.Context) error {
				calls.Add(1)
				return errors.New("callback must not run")
			}, func() error { releases.Add(1); return errors.New("release must not run") })
			if err := waitResult(t, host); err == nil || calls.Load() != 0 || releases.Load() != 0 {
				t.Fatalf("Run = %v, calls = %d, releases = %d; want rejection, 0, 0", err, calls.Load(), releases.Load())
			}
		})
	}
}

func TestRunInstancesHaveIndependentLifetimes(t *testing.T) {
	ctxA, cancelA := context.WithCancel(context.Background())
	ctxB, cancelB := context.WithCancel(context.Background())
	startedA := make(chan struct{}, 1)
	startsB := make(chan struct{}, 4)
	finishFirstB := make(chan struct{})
	var callsB, releasesA, releasesB atomic.Int32
	hostA := startHost(t, ctxA, cancelA, time.Hour, func(ctx context.Context) error {
		startedA <- struct{}{}
		<-ctx.Done()
		return nil
	}, func() error { releasesA.Add(1); return nil })
	hostB := startHost(t, ctxB, cancelB, time.Millisecond, func(ctx context.Context) error {
		call := callsB.Add(1)
		select {
		case startsB <- struct{}{}:
		case <-ctx.Done():
			return nil
		}
		if call == 1 {
			select {
			case <-finishFirstB:
			case <-ctx.Done():
			}
			return nil
		}
		cancelB()
		return nil
	}, func() error { releasesB.Add(1); return nil })
	waitFor(t, "first host start", startedA)
	waitFor(t, "second host start", startsB)
	cancelA()
	if err := waitResult(t, hostA); err != nil {
		t.Fatalf("first Run = %v; want nil", err)
	}
	select {
	case <-hostB.done:
		t.Fatalf("second host stopped when first host stopped: %v", hostB.err)
	default:
	}
	close(finishFirstB)
	waitFor(t, "second host recurrence after first host joined", startsB)
	if err := waitResult(t, hostB); err != nil || callsB.Load() != 2 || releasesA.Load() != 1 || releasesB.Load() != 1 {
		t.Fatalf("second Run = %v, calls = %d, releases = %d/%d; want nil, 2, 1/1", err, callsB.Load(), releasesA.Load(), releasesB.Load())
	}
}
