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

var _ func(context.Context, time.Duration, func(context.Context) error, func() error) error = indexer.Serve

func TestServeInvalidIntervalDoesNothing(t *testing.T) {
	t.Parallel()
	for _, interval := range []time.Duration{0, -time.Nanosecond} {
		err := indexer.Serve(context.Background(), interval,
			func(context.Context) error { t.Error("refresh called"); return nil },
			func() error { t.Error("release called"); return nil },
		)
		if !errors.Is(err, indexer.ErrInvalidInterval) {
			t.Fatalf("Serve(%s) = %v", interval, err)
		}
	}
}

func TestServeAlreadyCanceledReleasesOnceWithoutRefresh(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	releases := 0
	err := indexer.Serve(ctx, time.Millisecond,
		func(context.Context) error { t.Error("refresh started"); return nil },
		func() error { releases++; return nil },
	)
	if err != nil || releases != 1 {
		t.Fatalf("error/releases = %v/%d", err, releases)
	}
}

func TestServeWorkAndReleaseErrors(t *testing.T) {
	workErr, releaseErr := errors.New("refresh failed"), errors.New("release failed")
	for _, tc := range []struct {
		name          string
		work, release error
	}{
		{"success", nil, nil}, {"work only", workErr, nil}, {"release only", nil, releaseErr}, {"both", workErr, releaseErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls, releases := 0, 0
			completed := false
			err := indexer.Serve(ctx, time.Millisecond,
				func(got context.Context) error {
					calls++
					if got != ctx {
						t.Error("caller context changed")
					}
					cancel()
					completed = true
					return tc.work
				},
				func() error {
					releases++
					if !completed {
						t.Error("release overtook callback")
					}
					return tc.release
				},
			)
			if tc.work == nil && tc.release == nil && err != nil {
				t.Fatalf("Serve = %v; want nil", err)
			}
			for _, cause := range []error{tc.work, tc.release} {
				if cause != nil && !errors.Is(err, cause) {
					t.Fatalf("Serve = %v; missing %v", err, cause)
				}
			}
			if calls != 1 || releases != 1 {
				t.Fatalf("calls/releases = %d/%d", calls, releases)
			}
		})
	}
}

func TestServeWaitsForCancellationCleanup(t *testing.T) {
	for _, tc := range []struct {
		name        string
		callbackErr error
	}{
		{"successful cleanup", nil}, {"independent callback error", errors.New("cleanup failed")}, {"callback cancellation error", context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			started, cleaning, complete, released := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
			gate := make(chan struct{})
			var unlock sync.Once
			result := make(chan error, 1)
			finished := make(chan struct{})
			t.Cleanup(func() { cancel(); unlock.Do(func() { close(gate) }); awaitEvent(t, finished, "Serve cleanup join") })
			go func() {
				defer close(finished)
				result <- indexer.Serve(ctx, time.Millisecond,
					func(got context.Context) error {
						if got != ctx {
							return errors.New("caller context changed")
						}
						close(started)
						<-got.Done()
						close(cleaning)
						<-gate
						close(complete)
						return tc.callbackErr
					},
					func() error {
						select {
						case <-complete:
						default:
							return errors.New("release before callback cleanup")
						}
						close(released)
						return nil
					},
				)
			}()
			awaitEvent(t, started, "refresh startup")
			cancel()
			awaitEvent(t, cleaning, "cancellation cleanup startup")
			select {
			case <-released:
				t.Fatal("release overtook blocked cleanup")
			case err := <-result:
				t.Fatalf("Serve returned during cleanup: %v", err)
			case <-time.After(25 * time.Millisecond):
			}
			unlock.Do(func() { close(gate) })
			err := awaitResult(t, result)
			if tc.callbackErr == nil && err != nil {
				t.Fatalf("ordinary cancellation = %v; want nil", err)
			}
			if tc.callbackErr != nil && !errors.Is(err, tc.callbackErr) {
				t.Fatalf("callback error = %v; want %v", err, tc.callbackErr)
			}
			awaitEvent(t, released, "resource release")
			awaitEvent(t, finished, "Serve return")
		})
	}
}

func TestServeRecursAfterCompletionWithoutOverlap(t *testing.T) {
	t.Parallel()
	interval := 40 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan time.Time, 2)
	completed := make(chan time.Time, 1)
	gate := make(chan struct{})
	var unlock sync.Once
	var active atomic.Int32
	var calls, releases atomic.Int32
	done, finished := make(chan error, 1), make(chan struct{})
	t.Cleanup(func() {
		cancel()
		unlock.Do(func() { close(gate) })
		awaitEvent(t, finished, "recurrence cleanup join")
	})
	go func() {
		defer close(finished)
		done <- indexer.Serve(ctx, interval,
			func(got context.Context) error {
				if active.Add(1) != 1 {
					return errors.New("callbacks overlapped")
				}
				defer active.Add(-1)
				call := calls.Add(1)
				started <- time.Now()
				if got != ctx {
					return errors.New("caller context changed")
				}
				if call == 1 {
					<-gate
					completed <- time.Now()
					return nil
				}
				cancel()
				return nil
			},
			func() error { releases.Add(1); return nil },
		)
	}()
	first := awaitTime(t, started, "first callback")
	// Deliberately hold callback work longer than the interval; a ticker that
	// starts before completion would run its next callback immediately.
	timer := time.NewTimer(2 * interval)
	defer timer.Stop()
	<-timer.C
	unlock.Do(func() { close(gate) })
	completion := awaitTime(t, completed, "first callback completion")
	next := awaitTime(t, started, "second callback")
	if completion.Sub(first) < interval {
		t.Fatal("fixture did not hold first callback past interval")
	}
	if elapsed := next.Sub(completion); elapsed < interval-2*time.Millisecond {
		t.Fatalf("next callback after %s; want at least %s from completion", elapsed, interval)
	}
	if err := awaitResult(t, done); err != nil {
		t.Fatal(err)
	}
	awaitEvent(t, finished, "recurrence join")
	if calls.Load() != 2 || releases.Load() != 1 {
		t.Fatalf("calls/releases = %d/%d", calls.Load(), releases.Load())
	}
}

func TestServeIndependentInvocations(t *testing.T) {
	t.Parallel()
	type invocation struct {
		cancel   context.CancelFunc
		calls    chan int
		result   chan error
		finished chan struct{}
	}
	start := func() invocation {
		ctx, cancel := context.WithCancel(context.Background())
		run := invocation{cancel: cancel, calls: make(chan int, 32), result: make(chan error, 1), finished: make(chan struct{})}
		t.Cleanup(func() { cancel(); awaitEvent(t, run.finished, "independent invocation cleanup join") })
		go func() {
			defer close(run.finished)
			count := 0
			run.result <- indexer.Serve(ctx, 10*time.Millisecond,
				func(got context.Context) error {
					if got != ctx {
						return errors.New("independent context changed")
					}
					count++
					select {
					case run.calls <- count:
					case <-got.Done():
					}
					return nil
				},
				func() error { return nil },
			)
		}()
		return run
	}
	first, second := start(), start()
	awaitCount(t, first.calls, "first startup")
	awaitCount(t, second.calls, "second startup")
	first.cancel()
	if err := awaitResult(t, first.result); err != nil {
		t.Fatal(err)
	}
	awaitEvent(t, first.finished, "first join")
	// A fresh cycle after the first has joined proves the second still owns its
	// lifecycle. Drain queued cycles, then observe a later one.
	for draining := true; draining; {
		select {
		case <-second.calls:
		default:
			draining = false
		}
	}
	if n := awaitCount(t, second.calls, "second cycle after first joins"); n < 2 {
		t.Fatalf("second callback count = %d", n)
	}
	second.cancel()
	if err := awaitResult(t, second.result); err != nil {
		t.Fatal(err)
	}
	awaitEvent(t, second.finished, "second join")
}

func awaitEvent(t *testing.T, ch <-chan struct{}, label string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %s", label)
	}
}
func awaitResult(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for worker result")
		return nil
	}
}
func awaitTime(t *testing.T, ch <-chan time.Time, label string) time.Time {
	t.Helper()
	select {
	case stamp := <-ch:
		return stamp
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %s", label)
		return time.Time{}
	}
}
func awaitCount(t *testing.T, ch <-chan int, label string) int {
	t.Helper()
	select {
	case n := <-ch:
		return n
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %s", label)
		return 0
	}
}
