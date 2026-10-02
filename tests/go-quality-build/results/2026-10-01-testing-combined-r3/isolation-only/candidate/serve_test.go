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

const waitBound = 3 * time.Second

func receive[T any](t *testing.T, channel <-chan T, description string) T {
	t.Helper()
	select {
	case value := <-channel:
		return value
	case <-time.After(waitBound):
		t.Fatalf("timed out waiting for %s", description)
		var zero T
		return zero
	}
}

func TestServeInvalidAndAlreadyCanceled(t *testing.T) {
	t.Parallel()
	for _, interval := range []time.Duration{0, -time.Second} {
		calls, releases := 0, 0
		err := indexer.Serve(context.Background(), interval, func(context.Context) error { calls++; return nil }, func() error { releases++; return nil })
		if !errors.Is(err, indexer.ErrInvalidInterval) || calls != 0 || releases != 0 {
			t.Fatalf("invalid interval: %v, calls %d, releases %d", err, calls, releases)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls, releases := 0, 0
	releaseErr := errors.New("release failed")
	err := indexer.Serve(ctx, time.Millisecond, func(context.Context) error { calls++; return nil }, func() error { releases++; return releaseErr })
	if !errors.Is(err, releaseErr) || calls != 0 || releases != 1 {
		t.Fatalf("already canceled: %v, calls %d, releases %d", err, calls, releases)
	}
}

func TestServeFailureIdentities(t *testing.T) {
	t.Parallel()
	callbackErr := errors.New("refresh failed")
	releaseErr := errors.New("release failed")
	calls, releases := 0, 0
	err := indexer.Serve(context.Background(), time.Second, func(context.Context) error { calls++; return callbackErr }, func() error { releases++; return releaseErr })
	if !errors.Is(err, callbackErr) || !errors.Is(err, releaseErr) || calls != 1 || releases != 1 {
		t.Fatalf("failure: %v, calls %d, releases %d", err, calls, releases)
	}
}

func TestServeRecursFromCompletionWithoutOverlap(t *testing.T) {
	t.Parallel()
	interval := 60 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	gate := make(chan struct{})
	var gateOnce sync.Once
	openGate := func() { gateOnce.Do(func() { close(gate) }) }
	first := make(chan struct{}, 1)
	completed := make(chan time.Time, 1)
	second := make(chan time.Time, 1)
	result := make(chan error, 1)
	joined := make(chan struct{})
	var active atomic.Int32
	var releases atomic.Int32
	var count atomic.Int32
	overlap := errors.New("callbacks overlapped")
	t.Cleanup(func() {
		cancel()
		openGate()
		select {
		case <-joined:
		case <-time.After(waitBound):
			t.Error("Serve did not join")
		}
	})
	go func() {
		defer close(joined)
		result <- indexer.Serve(ctx, interval, func(got context.Context) error {
			if got != ctx {
				return errors.New("caller context not forwarded")
			}
			if active.Add(1) != 1 {
				active.Add(-1)
				return overlap
			}
			defer active.Add(-1)
			if count.Add(1) == 1 {
				first <- struct{}{}
				select {
				case <-gate:
				case <-ctx.Done():
					return ctx.Err()
				}
				completed <- time.Now()
				return nil
			}
			second <- time.Now()
			cancel()
			return nil
		}, func() error { releases.Add(1); return nil })
	}()
	receive(t, first, "immediate first callback")
	// This wait deliberately holds the callback beyond the requested interval.
	// It is not a readiness guess: the first callback has signaled entry.
	hold := time.NewTimer(2 * interval)
	defer hold.Stop()
	select {
	case <-hold.C:
	case err := <-result:
		t.Fatalf("Serve returned during callback: %v", err)
	}
	openGate()
	finished := receive(t, completed, "first callback completion")
	next := receive(t, second, "second callback")
	if elapsed := next.Sub(finished); elapsed < interval-5*time.Millisecond {
		t.Fatalf("next callback after %s; want at least %s after completion", elapsed, interval)
	}
	if err := receive(t, result, "Serve return"); err != nil {
		t.Fatal(err)
	}
	if count.Load() != 2 || releases.Load() != 1 || active.Load() != 0 {
		t.Fatalf("calls=%d releases=%d active=%d", count.Load(), releases.Load(), active.Load())
	}
}

func TestServeCancellationWaitsForCleanup(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"nil", "context_error", "wrapped_context_error", "parent_cause", "independent_error", "joined_context_error"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			if name == "parent_cause" {
				cancel()
				caused, cancelCause := context.WithCancelCause(context.Background())
				ctx, cancel = caused, func() { cancelCause(errors.New("parent stopping")) }
			}
			started := make(chan struct{})
			cleaning := make(chan struct{})
			cleanupGate := make(chan struct{})
			var once sync.Once
			finishCleanup := func() { once.Do(func() { close(cleanupGate) }) }
			released := make(chan struct{})
			joined := make(chan struct{})
			result := make(chan error, 1)
			callbackErr := errors.New("cleanup failed")
			releaseErr := errors.New("release failed")
			var releaseCount atomic.Int32
			t.Cleanup(func() {
				cancel()
				finishCleanup()
				select {
				case <-joined:
				case <-time.After(waitBound):
					t.Error("Serve cleanup did not join")
				}
			})
			go func() {
				defer close(joined)
				result <- indexer.Serve(ctx, time.Hour, func(got context.Context) error {
					if got != ctx {
						return errors.New("wrong callback context")
					}
					close(started)
					<-got.Done()
					close(cleaning)
					<-cleanupGate
					switch name {
					case "context_error":
						return got.Err()
					case "wrapped_context_error":
						return fmt.Errorf("cleanup: %w", got.Err())
					case "parent_cause":
						return fmt.Errorf("cleanup: %w", context.Cause(got))
					case "joined_context_error":
						return errors.Join(got.Err(), callbackErr)
					case "independent_error":
						return callbackErr
					default:
						return nil
					}
				}, func() error {
					releaseCount.Add(1)
					close(released)
					if name == "independent_error" || name == "joined_context_error" {
						return releaseErr
					}
					return nil
				})
			}()
			receive(t, started, "callback start")
			cancel()
			receive(t, cleaning, "callback cancellation cleanup")
			observation := time.NewTimer(25 * time.Millisecond)
			defer observation.Stop()
			select {
			case <-released:
				t.Fatal("release overtook callback cleanup")
			case <-observation.C:
			}
			finishCleanup()
			err := receive(t, result, "joined callback return")
			if name == "independent_error" || name == "joined_context_error" {
				if !errors.Is(err, callbackErr) || !errors.Is(err, releaseErr) {
					t.Fatalf("error identities = %v", err)
				}
			} else if err != nil {
				t.Fatalf("ordinary cancellation = %v", err)
			}
			receive(t, released, "release")
			if releaseCount.Load() != 1 {
				t.Fatalf("releases = %d", releaseCount.Load())
			}
		})
	}
}

func TestServeCancellationDuringInterval(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	called := make(chan struct{})
	joined := make(chan struct{})
	result := make(chan error, 1)
	var count, releases atomic.Int32
	t.Cleanup(func() {
		cancel()
		select {
		case <-joined:
		case <-time.After(waitBound):
			t.Error("Serve did not join")
		}
	})
	go func() {
		defer close(joined)
		result <- indexer.Serve(ctx, time.Hour, func(context.Context) error { count.Add(1); close(called); return nil }, func() error { releases.Add(1); return nil })
	}()
	receive(t, called, "first callback")
	cancel()
	if err := receive(t, result, "timer cancellation"); err != nil {
		t.Fatal(err)
	}
	if count.Load() != 1 || releases.Load() != 1 {
		t.Fatalf("calls=%d releases=%d", count.Load(), releases.Load())
	}
}

func TestServeIndependentInvocations(t *testing.T) {
	t.Parallel()
	type invocation struct {
		cancel   context.CancelFunc
		events   chan struct{}
		result   chan error
		joined   chan struct{}
		releases atomic.Int32
	}
	start := func() *invocation {
		ctx, cancel := context.WithCancel(context.Background())
		run := &invocation{cancel: cancel, events: make(chan struct{}, 20), result: make(chan error, 1), joined: make(chan struct{})}
		t.Cleanup(func() {
			run.cancel()
			select {
			case <-run.joined:
			case <-time.After(waitBound):
				t.Error("independent Serve did not join")
			}
		})
		go func() {
			defer close(run.joined)
			run.result <- indexer.Serve(ctx, 20*time.Millisecond, func(context.Context) error {
				select {
				case run.events <- struct{}{}:
				case <-ctx.Done():
				}
				return nil
			}, func() error { run.releases.Add(1); return nil })
		}()
		return run
	}
	first, second := start(), start()
	receive(t, first.events, "first invocation start")
	receive(t, second.events, "second invocation start")
	first.cancel()
	if err := receive(t, first.result, "first invocation stop"); err != nil {
		t.Fatal(err)
	}
	// Discard events already queued before the first invocation joined.
	for len(second.events) > 0 {
		<-second.events
	}
	receive(t, second.events, "second invocation recurrence after first stop")
	if first.releases.Load() != 1 || second.releases.Load() != 0 {
		t.Fatalf("releases: first=%d second=%d", first.releases.Load(), second.releases.Load())
	}
	second.cancel()
	if err := receive(t, second.result, "second invocation stop"); err != nil {
		t.Fatal(err)
	}
	if second.releases.Load() != 1 {
		t.Fatalf("second releases = %d", second.releases.Load())
	}
}
