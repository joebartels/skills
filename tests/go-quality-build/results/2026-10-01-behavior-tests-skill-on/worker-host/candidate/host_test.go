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

// Preserve callers that store Run as a function value, as well as direct calls.
var _ func(context.Context, time.Duration, func(context.Context) error, func() error) error = sweeper.Run

const waitLimit = 3 * time.Second

type runResult struct {
	done chan struct{}
	err  error
}

func startRun(t *testing.T, parent context.Context, interval time.Duration, sweep func(context.Context) error, release func() error) (context.CancelFunc, *runResult) {
	t.Helper()
	ctx, cancel := context.WithCancel(parent)
	r := &runResult{done: make(chan struct{})}
	go func() {
		r.err = sweeper.Run(ctx, interval, sweep, release)
		close(r.done)
	}()
	t.Cleanup(func() {
		cancel()
		timer := time.NewTimer(waitLimit)
		defer timer.Stop()
		select {
		case <-r.done:
		case <-timer.C:
			t.Error("Run did not finish during test cleanup")
		}
	})
	return cancel, r
}

func receive[T any](t *testing.T, ch <-chan T, what string) T {
	t.Helper()
	timer := time.NewTimer(waitLimit)
	defer timer.Stop()
	select {
	case value := <-ch:
		return value
	case <-timer.C:
		t.Fatalf("timed out waiting for %s", what)
		var zero T
		return zero
	}
}

func finished(t *testing.T, r *runResult) error {
	t.Helper()
	receive(t, r.done, "Run to return")
	return r.err
}

func TestInvalidIntervalHasNoEffects(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Nanosecond} {
		t.Run(interval.String(), func(t *testing.T) {
			var sweeps, releases atomic.Int32
			_, r := startRun(t, context.Background(), interval, func(context.Context) error {
				sweeps.Add(1)
				return errors.New("unexpected sweep")
			}, func() error {
				releases.Add(1)
				return nil
			})
			if err := finished(t, r); err == nil {
				t.Error("Run accepted a nonpositive interval")
			}
			if got := sweeps.Load(); got != 0 {
				t.Errorf("sweep calls = %d, want 0", got)
			}
			if got := releases.Load(); got != 0 {
				t.Errorf("release calls = %d, want 0", got)
			}
		})
	}
}

func TestAlreadyCanceledSkipsSweepAndReleases(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var sweeps, releases atomic.Int32
	_, r := startRun(t, ctx, time.Hour, func(context.Context) error {
		sweeps.Add(1)
		return errors.New("unexpected sweep")
	}, func() error {
		releases.Add(1)
		return nil
	})
	if err := finished(t, r); err != nil {
		t.Errorf("Run = %v, want nil", err)
	}
	if got := sweeps.Load(); got != 0 {
		t.Errorf("sweep calls = %d, want 0", got)
	}
	if got := releases.Load(); got != 1 {
		t.Errorf("release calls = %d, want 1", got)
	}
}

func TestImmediateSweepAndCancellationDuringInterval(t *testing.T) {
	started := make(chan context.Context, 2)
	var sweeps, releases atomic.Int32
	cancel, r := startRun(t, context.Background(), time.Hour, func(ctx context.Context) error {
		sweeps.Add(1)
		started <- ctx
		return nil
	}, func() error {
		releases.Add(1)
		return nil
	})
	workerCtx := receive(t, started, "immediate sweep")
	cancel()
	if err := finished(t, r); err != nil {
		t.Errorf("Run = %v, want nil", err)
	}
	if got := sweeps.Load(); got != 1 {
		t.Errorf("sweep calls = %d, want 1", got)
	}
	if got := releases.Load(); got != 1 {
		t.Errorf("release calls = %d, want 1", got)
	}
	if workerCtx.Err() != context.Canceled {
		t.Errorf("worker context = %v, want canceled", workerCtx.Err())
	}
}

func TestSuccessfulCyclesWaitAfterCompletionAndNeverOverlap(t *testing.T) {
	const interval = 60 * time.Millisecond
	workErr := errors.New("third cycle failed")
	type start struct {
		cycle  int32
		at     time.Time
		active int32
	}
	starts := make(chan start, 8)
	completions := make(chan time.Time, 8)
	firstMayFinish := make(chan struct{})
	var unblock sync.Once
	var sweeps, active, releases atomic.Int32
	var releasedWhileActive atomic.Bool
	_, r := startRun(t, context.Background(), interval, func(ctx context.Context) error {
		cycle := sweeps.Add(1)
		inFlight := active.Add(1)
		defer active.Add(-1)
		starts <- start{cycle: cycle, at: time.Now(), active: inFlight}
		if cycle == 1 {
			select {
			case <-firstMayFinish:
			case <-ctx.Done():
				return nil
			}
		}
		if cycle >= 3 {
			return workErr
		}
		completions <- time.Now()
		return nil
	}, func() error {
		if active.Load() != 0 {
			releasedWhileActive.Store(true)
		}
		releases.Add(1)
		return nil
	})
	t.Cleanup(func() { unblock.Do(func() { close(firstMayFinish) }) })
	first := receive(t, starts, "first cycle")
	if first.cycle != 1 || first.active != 1 {
		t.Fatalf("first cycle = %+v", first)
	}
	// Keep work running beyond an interval so a ticker or concurrent scheduler
	// cannot hide behind a quick callback.
	timer := time.NewTimer(2 * interval)
	select {
	case next := <-starts:
		timer.Stop()
		t.Fatalf("callback overlapped the blocked first cycle: %+v", next)
	case <-r.done:
		timer.Stop()
		t.Fatalf("Run returned while first callback was blocked: %v", r.err)
	case <-timer.C:
	}
	unblock.Do(func() { close(firstMayFinish) })
	completed := receive(t, completions, "first completion")
	for cycle := int32(2); cycle <= 3; cycle++ {
		next := receive(t, starts, "later cycle")
		if next.cycle != cycle || next.active != 1 {
			t.Errorf("cycle = %+v, want cycle %d with one active call", next, cycle)
		}
		if elapsed := next.at.Sub(completed); elapsed < interval {
			t.Errorf("cycle %d started %v after completion, want at least %v", cycle, elapsed, interval)
		}
		if cycle == 2 {
			completed = receive(t, completions, "second completion")
		}
	}
	if err := finished(t, r); !errors.Is(err, workErr) {
		t.Errorf("Run = %v, want third cycle error", err)
	}
	if got := sweeps.Load(); got != 3 {
		t.Errorf("sweep calls = %d, want 3", got)
	}
	if got := releases.Load(); got != 1 || releasedWhileActive.Load() {
		t.Errorf("release calls = %d, released while active = %v", got, releasedWhileActive.Load())
	}
}

func TestCancellationWaitsForCallbackCleanupBeforeReleaseAndReturn(t *testing.T) {
	workErr := errors.New("cleanup failed")
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "normal cancellation"},
		{name: "callback error during cancellation", err: workErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			started := make(chan struct{})
			cleaning := make(chan struct{})
			allowCleanup := make(chan struct{})
			released := make(chan struct{})
			var unblock sync.Once
			var cleanupDone, releaseBeforeCleanup atomic.Bool
			var sweeps, releases atomic.Int32
			cancel, r := startRun(t, context.Background(), time.Millisecond, func(ctx context.Context) error {
				sweeps.Add(1)
				close(started)
				<-ctx.Done()
				close(cleaning)
				<-allowCleanup
				cleanupDone.Store(true)
				return tc.err
			}, func() error {
				if !cleanupDone.Load() {
					releaseBeforeCleanup.Store(true)
				}
				releases.Add(1)
				close(released)
				return nil
			})
			t.Cleanup(func() { unblock.Do(func() { close(allowCleanup) }) })
			receive(t, started, "callback startup")
			cancel()
			receive(t, cleaning, "cancellation to reach callback cleanup")
			timer := time.NewTimer(30 * time.Millisecond)
			select {
			case <-released:
				timer.Stop()
				t.Fatal("release ran while callback cleanup was blocked")
			case <-r.done:
				timer.Stop()
				t.Fatal("Run returned while callback cleanup was blocked")
			case <-timer.C:
			}
			unblock.Do(func() { close(allowCleanup) })
			err := finished(t, r)
			if tc.err == nil && err != nil {
				t.Errorf("Run = %v, want nil", err)
			}
			if tc.err != nil && !errors.Is(err, tc.err) {
				t.Errorf("Run = %v, want callback cause %v", err, tc.err)
			}
			if got := sweeps.Load(); got != 1 {
				t.Errorf("sweep calls = %d, want 1", got)
			}
			if got := releases.Load(); got != 1 || releaseBeforeCleanup.Load() {
				t.Errorf("release calls = %d, release before cleanup = %v", got, releaseBeforeCleanup.Load())
			}
		})
	}
}

func TestWorkerAndReleaseErrorCombinations(t *testing.T) {
	workErr := errors.New("sweep failed")
	releaseErr := errors.New("release failed")
	for _, tc := range []struct {
		name             string
		work, releaseErr error
	}{
		{name: "normal cancellation"},
		{name: "sweep failure releases", work: workErr},
		{name: "release failure after cancellation", releaseErr: releaseErr},
		{name: "both fail", work: workErr, releaseErr: releaseErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var sweeps, releases atomic.Int32
			var releaseBeforeCancel atomic.Bool
			var workerCtx context.Context
			_, r := startRun(t, ctx, time.Hour, func(ctx context.Context) error {
				sweeps.Add(1)
				workerCtx = ctx
				if tc.work == nil {
					cancel()
				}
				return tc.work
			}, func() error {
				if workerCtx.Err() == nil {
					releaseBeforeCancel.Store(true)
				}
				releases.Add(1)
				return tc.releaseErr
			})
			err := finished(t, r)
			if tc.work == nil && tc.releaseErr == nil && err != nil {
				t.Errorf("Run = %v, want nil", err)
			}
			for _, want := range []error{tc.work, tc.releaseErr} {
				if want != nil && !errors.Is(err, want) {
					t.Errorf("Run = %v, want cause %v", err, want)
				}
			}
			if tc.work != nil && ctx.Err() != nil {
				t.Errorf("Run canceled its caller's context: %v", ctx.Err())
			}
			if got := sweeps.Load(); got != 1 {
				t.Errorf("sweep calls = %d, want 1", got)
			}
			if got := releases.Load(); got != 1 || releaseBeforeCancel.Load() {
				t.Errorf("release calls = %d, release before cancel = %v", got, releaseBeforeCancel.Load())
			}
		})
	}
}

func TestRunWaitsForReleaseBeforeReturning(t *testing.T) {
	workErr := errors.New("sweep failed")
	releaseErr := errors.New("release failed")
	releasing := make(chan struct{})
	allowRelease := make(chan struct{})
	var unblock sync.Once
	_, r := startRun(t, context.Background(), time.Hour, func(context.Context) error {
		return workErr
	}, func() error {
		close(releasing)
		<-allowRelease
		return releaseErr
	})
	t.Cleanup(func() { unblock.Do(func() { close(allowRelease) }) })
	receive(t, releasing, "release startup")
	timer := time.NewTimer(30 * time.Millisecond)
	select {
	case <-r.done:
		timer.Stop()
		t.Fatal("Run returned while release was blocked")
	case <-timer.C:
	}
	unblock.Do(func() { close(allowRelease) })
	if err := finished(t, r); !errors.Is(err, workErr) || !errors.Is(err, releaseErr) {
		t.Errorf("Run = %v, want worker and release causes", err)
	}
}

func TestRunsHaveIndependentLifetimes(t *testing.T) {
	firstStarted := make(chan int32, 8)
	secondStarted := make(chan int32, 8)
	var firstCalls, secondCalls, firstReleases, secondReleases atomic.Int32
	callback := func(calls *atomic.Int32, starts chan<- int32) func(context.Context) error {
		return func(ctx context.Context) error {
			call := calls.Add(1)
			select {
			case starts <- call:
			case <-ctx.Done():
			}
			return nil
		}
	}
	cancelFirst, first := startRun(t, context.Background(), 20*time.Millisecond, callback(&firstCalls, firstStarted), func() error {
		firstReleases.Add(1)
		return nil
	})
	cancelSecond, second := startRun(t, context.Background(), 20*time.Millisecond, callback(&secondCalls, secondStarted), func() error {
		secondReleases.Add(1)
		return nil
	})
	receive(t, firstStarted, "first worker startup")
	receive(t, secondStarted, "second worker startup")
	cancelFirst()
	if err := finished(t, first); err != nil {
		t.Errorf("first Run = %v, want nil", err)
	}
	before := secondCalls.Load()
	for receive(t, secondStarted, "second worker cycle after first joined") <= before {
	}
	select {
	case <-second.done:
		t.Fatalf("second Run stopped with the first: %v", second.err)
	default:
	}
	cancelSecond()
	if err := finished(t, second); err != nil {
		t.Errorf("second Run = %v, want nil", err)
	}
	if firstReleases.Load() != 1 || secondReleases.Load() != 1 {
		t.Errorf("release calls = (%d, %d), want (1, 1)", firstReleases.Load(), secondReleases.Load())
	}
}
