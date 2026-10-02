package indexer_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"example.com/indexer"
)

func TestServeRejectsIntervalBeforeOwnership(t *testing.T) {
	t.Parallel()
	for _, interval := range []time.Duration{0, -time.Second} {
		calls, releases := 0, 0
		err := indexer.Serve(context.Background(), interval, func(context.Context) error { calls++; return nil }, func() error { releases++; return nil })
		if !errors.Is(err, indexer.ErrInvalidInterval) || calls != 0 || releases != 0 {
			t.Fatalf("Serve = %v, calls=%d, releases=%d", err, calls, releases)
		}
	}
}

func TestServeAlreadyCanceledReleasesWithoutRefresh(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	releaseErr := errors.New("release failed")
	calls, releases := 0, 0
	err := indexer.Serve(ctx, time.Millisecond, func(context.Context) error { calls++; return nil }, func() error { releases++; return releaseErr })
	if !errors.Is(err, releaseErr) || calls != 0 || releases != 1 {
		t.Fatalf("Serve = %v, calls=%d, releases=%d", err, calls, releases)
	}
}

func TestServeJoinsCallbackAndReleaseErrors(t *testing.T) {
	t.Parallel()
	callbackErr := errors.New("refresh failed")
	releaseErr := errors.New("release failed")
	calls, releases := 0, 0
	err := indexer.Serve(context.Background(), time.Millisecond, func(context.Context) error { calls++; return callbackErr }, func() error { releases++; return releaseErr })
	if !errors.Is(err, callbackErr) || !errors.Is(err, releaseErr) || calls != 1 || releases != 1 {
		t.Fatalf("Serve = %v, calls=%d, releases=%d", err, calls, releases)
	}
}

func TestServeCancellationWaitsForCallbackCleanup(t *testing.T) {
	t.Parallel()
	for _, failure := range []string{"none", "independent", "callback-context-error"} {
		t.Run(failure, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			started := make(chan struct{})
			cleaning := make(chan struct{})
			allowCleanup := make(chan struct{})
			released := make(chan struct{})
			finished := make(chan struct{})
			result := make(chan error, 1)
			var allowOnce sync.Once
			unblock := func() { allowOnce.Do(func() { close(allowCleanup) }) }
			failureErr := errors.New("independent cleanup failure")
			if failure == "callback-context-error" {
				failureErr = context.Canceled
			}
			if failure == "none" {
				failureErr = nil
			}
			t.Cleanup(func() {
				cancel()
				unblock()
				awaitEvent(t, finished, "Serve completion during cleanup")
			})
			go func() {
				defer close(finished)
				result <- indexer.Serve(ctx, time.Millisecond, func(got context.Context) error {
					if got != ctx {
						return errors.New("callback did not receive caller context")
					}
					close(started)
					<-got.Done()
					close(cleaning)
					<-allowCleanup
					return failureErr
				}, func() error { close(released); return nil })
			}()
			awaitEvent(t, started, "immediate refresh")
			cancel()
			awaitEvent(t, cleaning, "callback cancellation cleanup")
			select {
			case <-released:
				t.Fatal("release overtook blocked cleanup")
			default:
			}
			select {
			case err := <-result:
				t.Fatalf("Serve returned before cleanup: %v", err)
			default:
			}
			unblock()
			err := awaitError(t, result, "Serve after cleanup")
			if failureErr == nil && err != nil {
				t.Fatalf("ordinary cancellation = %v; want nil", err)
			}
			if failureErr != nil && !errors.Is(err, failureErr) {
				t.Fatalf("Serve = %v; want callback error identity", err)
			}
			awaitEvent(t, released, "release after callback cleanup")
		})
	}
}

func TestServeWaitsIntervalFromCompletionAndDoesNotOverlap(t *testing.T) {
	t.Parallel()
	const interval = 40 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	starts := make(chan time.Time, 3)
	completed := make(chan time.Time, 1)
	allowFirst := make(chan struct{})
	finished := make(chan struct{})
	result := make(chan error, 1)
	var allowOnce sync.Once
	unblock := func() { allowOnce.Do(func() { close(allowFirst) }) }
	calls, releases := 0, 0 // Read only after the Serve goroutine has joined.
	t.Cleanup(func() { cancel(); unblock(); awaitEvent(t, finished, "recurring Serve cleanup") })
	go func() {
		defer close(finished)
		result <- indexer.Serve(ctx, interval, func(context.Context) error {
			calls++
			starts <- time.Now()
			if calls == 1 {
				<-allowFirst
				completed <- time.Now()
				return nil
			}
			cancel()
			return nil
		}, func() error { releases++; return nil })
	}()
	awaitTime(t, starts, "first callback")
	// Hold the callback longer than the interval to expose tick-based scheduling.
	wait := time.NewTimer(2 * interval)
	defer wait.Stop()
	select {
	case <-wait.C:
	case next := <-starts:
		t.Fatalf("overlapping callback started at %v", next)
	}
	unblock()
	completion := awaitTime(t, completed, "first callback completion")
	second := awaitTime(t, starts, "second callback")
	// The completion event is emitted just before return, so it provides a
	// conservative lower bound; 2ms allows timer/platform granularity.
	if elapsed := second.Sub(completion); elapsed < interval-2*time.Millisecond {
		t.Fatalf("next callback after %v; want at least %v from completion", elapsed, interval)
	}
	if err := awaitError(t, result, "recurring Serve cancellation"); err != nil {
		t.Fatal(err)
	}
	awaitEvent(t, finished, "recurring Serve join")
	if calls != 2 || releases != 1 {
		t.Fatalf("calls=%d releases=%d; want 2 and 1", calls, releases)
	}
}

func TestServeCancellationDuringInterval(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	returned := make(chan struct{})
	finished := make(chan struct{})
	result := make(chan error, 1)
	calls, releases := 0, 0
	t.Cleanup(func() { cancel(); awaitEvent(t, finished, "interval cancellation cleanup") })
	go func() {
		defer close(finished)
		result <- indexer.Serve(ctx, time.Hour, func(context.Context) error { calls++; close(returned); return nil }, func() error { releases++; return nil })
	}()
	awaitEvent(t, returned, "successful callback")
	cancel()
	if err := awaitError(t, result, "interval cancellation"); err != nil {
		t.Fatal(err)
	}
	awaitEvent(t, finished, "interval cancellation join")
	if calls != 1 || releases != 1 {
		t.Fatalf("calls=%d releases=%d; want 1 and 1", calls, releases)
	}
}

func TestServeInvocationsAreIndependent(t *testing.T) {
	t.Parallel()
	type ownedRun struct {
		cancel   context.CancelFunc
		starts   chan struct{}
		finished chan struct{}
		result   chan error
	}
	start := func() ownedRun {
		ctx, cancel := context.WithCancel(context.Background())
		run := ownedRun{cancel, make(chan struct{}, 32), make(chan struct{}), make(chan error, 1)}
		t.Cleanup(func() { cancel(); awaitEvent(t, run.finished, "independent run cleanup") })
		go func() {
			defer close(run.finished)
			run.result <- indexer.Serve(ctx, 10*time.Millisecond, func(context.Context) error {
				select {
				case run.starts <- struct{}{}:
				case <-ctx.Done():
				}
				return nil
			}, func() error { return nil })
		}()
		return run
	}
	first, second := start(), start()
	awaitEvent(t, first.starts, "first independent startup")
	awaitEvent(t, second.starts, "second independent startup")
	first.cancel()
	if err := awaitError(t, first.result, "first independent stop"); err != nil {
		t.Fatal(err)
	}
	awaitEvent(t, first.finished, "first independent join")
	// Drain pre-join events before observing new progress by the surviving run.
drain:
	for {
		select {
		case <-second.starts:
		default:
			break drain
		}
	}
	awaitEvent(t, second.starts, "second run progress after first joined")
	second.cancel()
	if err := awaitError(t, second.result, "second independent stop"); err != nil {
		t.Fatal(err)
	}
}

func awaitTime(t *testing.T, event <-chan time.Time, what string) time.Time {
	t.Helper()
	select {
	case when := <-event:
		return when
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
		return time.Time{}
	}
}
