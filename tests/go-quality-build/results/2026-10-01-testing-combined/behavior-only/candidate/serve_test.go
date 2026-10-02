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

func TestServeInvalidAndCanceledStartup(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Second} {
		calls, releases := 0, 0
		err := indexer.Serve(context.Background(), interval, func(context.Context) error { calls++; return nil }, func() error { releases++; return nil })
		if !errors.Is(err, indexer.ErrInvalidInterval) || calls != 0 || releases != 0 {
			t.Fatalf("interval %s: error = %v, calls = %d, releases = %d", interval, err, calls, releases)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls, releases := 0, 0
	err := indexer.Serve(ctx, time.Second, func(context.Context) error { calls++; return nil }, func() error { releases++; return nil })
	if err != nil || calls != 0 || releases != 1 {
		t.Fatalf("canceled startup: %v, calls = %d, releases = %d", err, calls, releases)
	}
}

func TestServeErrorCombinations(t *testing.T) {
	workErr, releaseErr := errors.New("work failed"), errors.New("release failed")
	for _, tc := range []struct {
		name          string
		work, release error
	}{
		{"success", nil, nil},
		{"work fails", workErr, nil},
		{"release fails", nil, releaseErr},
		{"both fail", workErr, releaseErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls, releases := 0, 0
			err := indexer.Serve(ctx, time.Millisecond, func(got context.Context) error {
				if got != ctx {
					t.Error("context was replaced")
				}
				calls++
				cancel()
				return tc.work
			}, func() error { releases++; return tc.release })
			if tc.work == nil && tc.release == nil && err != nil {
				t.Fatalf("error = %v, want nil", err)
			}
			for _, cause := range []error{tc.work, tc.release} {
				if cause != nil && !errors.Is(err, cause) {
					t.Errorf("error = %v, want cause %v", err, cause)
				}
			}
			if calls != 1 || releases != 1 {
				t.Fatalf("calls = %d, releases = %d", calls, releases)
			}
		})
	}
}

func TestServeWaitsForCancellationCleanup(t *testing.T) {
	for _, callbackErr := range []error{nil, errors.New("independent failure"), context.Canceled} {
		t.Run(errorName(callbackErr), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			entered, cleaning, allowCleanup, cleaned, released := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
			var unblock sync.Once
			t.Cleanup(func() { cancel(); unblock.Do(func() { close(allowCleanup) }) })
			done := make(chan error, 1)
			go func() {
				done <- indexer.Serve(ctx, time.Second, func(got context.Context) error {
					if got != ctx {
						t.Error("context was replaced")
					}
					close(entered)
					<-got.Done()
					close(cleaning)
					<-allowCleanup
					close(cleaned)
					return callbackErr
				}, func() error {
					select {
					case <-cleaned:
					default:
						t.Error("release overtook cleanup")
					}
					close(released)
					return nil
				})
			}()
			await(t, entered)
			cancel()
			await(t, cleaning)
			select {
			case <-released:
				t.Fatal("released while callback was cleaning")
			case err := <-done:
				t.Fatalf("returned before cleanup: %v", err)
			case <-time.After(30 * time.Millisecond):
			}
			unblock.Do(func() { close(allowCleanup) })
			err := await(t, done)
			if callbackErr == nil && err != nil {
				t.Fatalf("ordinary cancellation = %v, want nil", err)
			}
			if callbackErr != nil && !errors.Is(err, callbackErr) {
				t.Fatalf("error = %v, want callback cause", err)
			}
			await(t, released)
		})
	}
}

func TestServeIntervalStartsAfterCompletion(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	first, allowFirst, next := make(chan struct{}), make(chan struct{}), make(chan time.Time, 1)
	var unblock sync.Once
	t.Cleanup(func() { cancel(); unblock.Do(func() { close(allowFirst) }) })
	const interval = 80 * time.Millisecond
	done := make(chan error, 1)
	var active atomic.Int32
	calls, releases := 0, 0
	go func() {
		done <- indexer.Serve(ctx, interval, func(got context.Context) error {
			if got != ctx {
				t.Error("context was replaced")
			}
			if active.Add(1) != 1 {
				t.Error("overlapping refreshes")
			}
			defer active.Add(-1)
			calls++
			if calls == 1 {
				close(first)
				<-allowFirst
				return nil
			}
			if calls == 2 {
				next <- time.Now()
				cancel()
			}
			return nil
		}, func() error { releases++; return nil })
	}()
	await(t, first)
	// Hold the first callback longer than interval; a ticker would fire too soon.
	select {
	case <-next:
		t.Fatal("second callback overlapped first")
	case <-time.After(2 * interval):
	}
	completed := time.Now()
	unblock.Do(func() { close(allowFirst) })
	started := await(t, next)
	if elapsed := started.Sub(completed); elapsed < interval {
		t.Errorf("next invocation after %s, want at least %s from completion", elapsed, interval)
	}
	if err := await(t, done); err != nil {
		t.Fatal(err)
	}
	if calls != 2 || releases != 1 {
		t.Fatalf("calls = %d, releases = %d", calls, releases)
	}
}

func TestServeIndependentInvocations(t *testing.T) {
	ctxA, cancelA := context.WithCancel(context.Background())
	ctxB, cancelB := context.WithCancel(context.Background())
	t.Cleanup(func() { cancelA(); cancelB() })
	eventsA, eventsB := make(chan struct{}, 16), make(chan struct{}, 16)
	doneA, doneB := make(chan error, 1), make(chan error, 1)
	start := func(ctx context.Context, events chan struct{}, done chan error) {
		go func() {
			done <- indexer.Serve(ctx, 20*time.Millisecond, func(context.Context) error {
				select {
				case events <- struct{}{}:
				default:
				}
				return nil
			}, func() error { return nil })
		}()
	}
	start(ctxA, eventsA, doneA)
	start(ctxB, eventsB, doneB)
	await(t, eventsA)
	await(t, eventsB)
	cancelA()
	if err := await(t, doneA); err != nil {
		t.Fatal(err)
	}
	// Drain events already observed before A stopped, then require a fresh cycle.
	for len(eventsB) > 0 {
		<-eventsB
	}
	await(t, eventsB)
	select {
	case err := <-doneB:
		t.Fatalf("B stopped with A: %v", err)
	default:
	}
	cancelB()
	if err := await(t, doneB); err != nil {
		t.Fatal(err)
	}
}

func errorName(err error) string {
	if err == nil {
		return "nil"
	}
	return err.Error()
}
func await[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for lifecycle event")
	}
	var zero T
	return zero
}
