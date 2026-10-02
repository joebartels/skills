package indexer_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/indexer"
)

func TestServeInvalidIntervalHasNoLifecycle(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Second} {
		called := 0
		err := indexer.Serve(context.Background(), interval, func(context.Context) error { called++; return nil }, func() error { called++; return nil })
		if !errors.Is(err, indexer.ErrInvalidInterval) || called != 0 {
			t.Fatalf("Serve(%s) = %v, calls %d; want ErrInvalidInterval and no calls", interval, err, called)
		}
	}
}

func TestServeAlreadyCanceledReleasesWithoutRefresh(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	refreshes, releases := 0, 0
	err := indexer.Serve(ctx, time.Second, func(context.Context) error { refreshes++; return nil }, func() error { releases++; return nil })
	if err != nil || refreshes != 0 || releases != 1 {
		t.Fatalf("Serve = %v, refreshes %d, releases %d", err, refreshes, releases)
	}
}

func TestServeIndependentErrorCombinations(t *testing.T) {
	workErr := errors.New("refresh failed")
	releaseErr := errors.New("release failed")
	for _, tc := range []struct {
		name          string
		work, release error
	}{
		{"ordinary cancellation", nil, nil},
		{"refresh only", workErr, nil},
		{"release only", nil, releaseErr},
		{"both", workErr, releaseErr},
		{"callback cancellation error", context.Canceled, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			refreshes, releases := 0, 0
			err := indexer.Serve(ctx, time.Millisecond, func(got context.Context) error {
				if got != ctx {
					t.Error("callback did not receive caller context")
				}
				refreshes++
				if tc.work == nil {
					cancel()
				}
				return tc.work
			}, func() error { releases++; return tc.release })
			if tc.work == nil && tc.release == nil && err != nil {
				t.Fatalf("Serve = %v; want nil", err)
			}
			for _, cause := range []error{tc.work, tc.release} {
				if cause != nil && !errors.Is(err, cause) {
					t.Errorf("Serve = %v; missing cause %v", err, cause)
				}
			}
			if refreshes != 1 || releases != 1 {
				t.Fatalf("refreshes %d, releases %d; want one each", refreshes, releases)
			}
		})
	}
}

func TestServeWaitsFromCallbackCompletionAndNeverOverlaps(t *testing.T) {
	interval := 40 * time.Millisecond
	firstStarted := make(chan struct{})
	gate := make(chan struct{})
	var gateOnce sync.Once
	unblock := func() { gateOnce.Do(func() { close(gate) }) }
	completion := make(chan time.Time, 1)
	secondStarted := make(chan time.Time, 1)
	var calls, active atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	done, joined := launchServe(t, ctx, cancel, unblock, interval, func(got context.Context) error {
		if active.Add(1) != 1 {
			return errors.New("callbacks overlap")
		}
		defer active.Add(-1)
		n := calls.Add(1)
		if n == 1 {
			close(firstStarted)
			select {
			case <-gate:
				completion <- time.Now()
			case <-got.Done():
			}
			return nil
		}
		secondStarted <- time.Now()
		cancel()
		return nil
	}, func() error { return nil })
	awaitEvent(t, firstStarted, "first refresh")
	// Deliberately hold the first callback past the interval. A ticker based on
	// start time would have a ready tick when this callback finally returns.
	timer := time.NewTimer(2 * interval)
	defer timer.Stop()
	select {
	case <-timer.C:
	case err := <-done:
		t.Fatalf("Serve exited while callback was held: %v", err)
	}
	unblock()
	var completed, next time.Time
	select {
	case completed = <-completion:
	case <-time.After(5 * time.Second):
		t.Fatal("first callback did not complete")
	}
	select {
	case next = <-secondStarted:
	case <-time.After(5 * time.Second):
		select {
		case err := <-done:
			t.Fatalf("Serve exited before recurrence: %v", err)
		default:
			t.Fatal("second refresh did not start")
		}
	}
	if elapsed := next.Sub(completed); elapsed < interval {
		t.Errorf("delay after completion = %s; want at least %s", elapsed, interval)
	}
	awaitEvent(t, joined, "Serve completion")
	if err := awaitResult(t, done); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 || active.Load() != 0 {
		t.Fatalf("calls %d, active %d; want two completed callbacks", calls.Load(), active.Load())
	}
}

func TestServeCancellationJoinsCleanupBeforeRelease(t *testing.T) {
	for _, callbackErr := range []error{nil, errors.New("cleanup failed")} {
		name := "nil callback result"
		if callbackErr != nil {
			name = "callback error"
		}
		t.Run(name, func(t *testing.T) {
			started := make(chan struct{})
			cleanupStarted := make(chan struct{})
			gate := make(chan struct{})
			var gateOnce sync.Once
			unblock := func() { gateOnce.Do(func() { close(gate) }) }
			released := make(chan struct{})
			var finished atomic.Bool
			var releaseCount atomic.Int32
			ctx, cancel := context.WithCancel(context.Background())
			done, _ := launchServe(t, ctx, cancel, unblock, time.Hour, func(got context.Context) error {
				close(started)
				<-got.Done()
				close(cleanupStarted)
				<-gate
				finished.Store(true)
				return callbackErr
			}, func() error {
				releaseCount.Add(1)
				close(released)
				if !finished.Load() {
					return errors.New("release overtook cleanup")
				}
				return nil
			})
			awaitEvent(t, started, "callback startup")
			cancel()
			awaitEvent(t, cleanupStarted, "callback cancellation cleanup")
			select {
			case <-released:
				t.Fatal("release ran while callback cleanup was held")
			case err := <-done:
				t.Fatalf("Serve returned while cleanup was held: %v", err)
			case <-time.After(30 * time.Millisecond):
			}
			unblock()
			err := awaitResult(t, done)
			if callbackErr == nil && err != nil || callbackErr != nil && !errors.Is(err, callbackErr) {
				t.Fatalf("Serve = %v; want callback result %v", err, callbackErr)
			}
			awaitEvent(t, released, "resource release")
			if releaseCount.Load() != 1 {
				t.Fatalf("release calls = %d; want 1", releaseCount.Load())
			}
		})
	}
}

func TestServeIndependentInvocations(t *testing.T) {
	starts := []chan struct{}{make(chan struct{}, 8), make(chan struct{}, 8)}
	cancels := make([]context.CancelFunc, 2)
	results := make([]<-chan error, 2)
	var releases [2]atomic.Int32
	for n := range starts {
		ctx, cancel := context.WithCancel(context.Background())
		cancels[n] = cancel
		results[n], _ = launchServe(t, ctx, cancel, func() {}, 10*time.Millisecond, func(context.Context) error {
			select {
			case starts[n] <- struct{}{}:
			default:
			}
			return nil
		}, func() error { releases[n].Add(1); return nil })
	}
	awaitEvent(t, starts[0], "first instance startup")
	awaitEvent(t, starts[1], "second instance startup")
	cancels[0]()
	if err := awaitResult(t, results[0]); err != nil {
		t.Fatal(err)
	}
	// Discard already-buffered starts, then require a later start from the live
	// instance after the first invocation has returned.
	for len(starts[1]) > 0 {
		<-starts[1]
	}
	awaitEvent(t, starts[1], "second instance recurrence after first stopped")
	cancels[1]()
	if err := awaitResult(t, results[1]); err != nil {
		t.Fatal(err)
	}
	for n := range releases {
		if releases[n].Load() != 1 {
			t.Errorf("instance %d releases = %d", n, releases[n].Load())
		}
	}
}

func launchServe(t *testing.T, ctx context.Context, cancel context.CancelFunc, unblock func(), interval time.Duration, refresh func(context.Context) error, release func() error) (<-chan error, <-chan struct{}) {
	t.Helper()
	done := make(chan error, 1)
	joined := make(chan struct{})
	t.Cleanup(func() { cancel(); unblock(); awaitEvent(t, joined, "Serve cleanup join") })
	go func() {
		done <- indexer.Serve(ctx, interval, refresh, release)
		close(joined)
	}()
	return done, joined
}

func awaitEvent(t *testing.T, ch <-chan struct{}, label string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", label)
	}
}

func awaitResult(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for operation result")
		return nil
	}
}
