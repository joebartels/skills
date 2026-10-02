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

func TestServeInvalidAndCanceledStartup(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Second} {
		t.Run(interval.String(), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			calls, releases := 0, 0
			err := indexer.Serve(ctx, interval, func(context.Context) error { calls++; return nil }, func() error { releases++; return nil })
			if !errors.Is(err, indexer.ErrInvalidInterval) || calls != 0 || releases != 0 {
				t.Fatalf("Serve = %v; refreshes = %d, releases = %d", err, calls, releases)
			}
		})
	}
	t.Run("already canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		calls, releases := 0, 0
		err := indexer.Serve(ctx, time.Second, func(context.Context) error { calls++; return nil }, func() error { releases++; return nil })
		if err != nil || calls != 0 || releases != 1 {
			t.Fatalf("Serve = %v; refreshes = %d, releases = %d", err, calls, releases)
		}
	})
}

func TestServeErrorCombinations(t *testing.T) {
	workErr, releaseErr := errors.New("refresh failed"), errors.New("release failed")
	for _, tc := range []struct {
		name          string
		work, release error
	}{
		{"success", nil, nil},
		{"work only", workErr, nil},
		{"release only", nil, releaseErr},
		{"both", workErr, releaseErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls, releases := 0, 0
			err := indexer.Serve(ctx, time.Millisecond, func(got context.Context) error {
				calls++
				if got != ctx {
					t.Error("callback context was replaced")
				}
				if tc.work == nil {
					cancel()
				}
				return tc.work
			}, func() error { releases++; return tc.release })
			if tc.work == nil && tc.release == nil && err != nil {
				t.Fatalf("Serve = %v; want nil", err)
			}
			for _, want := range []error{tc.work, tc.release} {
				if want != nil && !errors.Is(err, want) {
					t.Errorf("Serve = %v; want cause %v", err, want)
				}
			}
			if calls != 1 || releases != 1 {
				t.Fatalf("refreshes = %d, releases = %d; want 1 each", calls, releases)
			}
		})
	}
}

// A valid error need not have a comparable dynamic value.
type sliceError struct{ values []string }

func (e sliceError) Error() string { return fmt.Sprint(e.values) }

func TestServeCancellationJoinsCleanup(t *testing.T) {
	independent := errors.New("cleanup also failed")
	for _, tc := range []struct {
		name string
		err  error
		want error
	}{
		{"nil", nil, nil},
		{"ordinary cancellation", fmt.Errorf("request: %w", context.Canceled), nil},
		{"joined independent failure", errors.Join(context.Canceled, independent), independent},
		{"noncomparable failure", sliceError{[]string{"failure"}}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			started, cleaning := make(chan struct{}), make(chan struct{})
			gate, finished := make(chan struct{}), make(chan struct{})
			released := make(chan bool, 1)
			result := make(chan error, 1)
			var unblock sync.Once
			var cleanupComplete atomic.Bool
			t.Cleanup(func() {
				cancel()
				unblock.Do(func() { close(gate) })
				waitClosed(t, finished, "Serve cleanup")
			})
			go func() {
				defer close(finished)
				result <- indexer.Serve(ctx, time.Millisecond, func(got context.Context) error {
					if got != ctx {
						return errors.New("callback context changed")
					}
					close(started)
					<-got.Done()
					close(cleaning)
					<-gate
					cleanupComplete.Store(true)
					return tc.err
				}, func() error { released <- cleanupComplete.Load(); return nil })
			}()
			waitClosed(t, started, "refresh start")
			cancel()
			waitClosed(t, cleaning, "cooperative cleanup start")
			select {
			case <-released:
				t.Fatal("release overtook callback cleanup")
			case <-finished:
				t.Fatal("Serve returned before callback cleanup")
			case <-time.After(20 * time.Millisecond):
			}
			unblock.Do(func() { close(gate) })
			waitClosed(t, finished, "Serve cancellation join")
			if complete := <-released; !complete {
				t.Fatal("release ran before callback finished cleanup")
			}
			err := <-result
			switch {
			case tc.want != nil:
				if !errors.Is(err, tc.want) || !errors.Is(err, context.Canceled) {
					t.Fatalf("Serve = %v; want independent and cancellation identities", err)
				}
			case tc.name == "noncomparable failure":
				var got sliceError
				if !errors.As(err, &got) || len(got.values) != 1 || got.values[0] != "failure" {
					t.Fatalf("Serve = %v; want sliceError", err)
				}
			default:
				if err != nil {
					t.Fatalf("Serve = %v; want nil", err)
				}
			}
		})
	}
}

func TestServeUnrelatedCancellationErrorIsFailure(t *testing.T) {
	calls := 0
	err := indexer.Serve(context.Background(), time.Second, func(context.Context) error { return context.Canceled }, func() error { calls++; return nil })
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("Serve = %v, releases = %d; want callback error and release", err, calls)
	}
}

func TestServeSequentialRecurrenceFromCompletion(t *testing.T) {
	const interval = 60 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan int, 4)
	gate, finished := make(chan struct{}), make(chan struct{})
	result := make(chan error, 1)
	var unblock sync.Once
	var completed, next time.Time
	var releases int
	t.Cleanup(func() {
		cancel()
		unblock.Do(func() { close(gate) })
		waitClosed(t, finished, "recurrence cleanup")
	})
	go func() {
		defer close(finished)
		calls := 0
		result <- indexer.Serve(ctx, interval, func(context.Context) error {
			calls++
			started <- calls
			if calls == 1 {
				<-gate
				completed = time.Now()
				return nil
			}
			next = time.Now()
			cancel()
			return nil
		}, func() error { releases++; return nil })
	}()
	select {
	case n := <-started:
		if n != 1 {
			t.Fatalf("first callback = %d", n)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("first callback did not start")
	}
	// Hold the first call for 2.5 intervals: a ticker-based implementation
	// cannot mistake this gate for completion or use a queued old tick.
	select {
	case n := <-started:
		t.Fatalf("callback %d overlapped the held first call", n)
	case <-time.After(5 * interval / 2):
	}
	unblock.Do(func() { close(gate) })
	waitClosed(t, finished, "later successful callback")
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	if delta := next.Sub(completed); delta < interval {
		t.Fatalf("next callback after %v; want at least %v from completion", delta, interval)
	}
	if releases != 1 {
		t.Fatalf("releases = %d; want 1", releases)
	}
}

func TestServeIndependentInvocations(t *testing.T) {
	type run struct {
		cancel   context.CancelFunc
		calls    chan struct{}
		finished chan struct{}
		result   chan error
		releases atomic.Int32
	}
	var runs [2]*run
	for n := range runs {
		ctx, cancel := context.WithCancel(context.Background())
		r := &run{cancel: cancel, calls: make(chan struct{}, 16), finished: make(chan struct{}), result: make(chan error, 1)}
		runs[n] = r
		t.Cleanup(func() { r.cancel(); waitClosed(t, r.finished, "independent run cleanup") })
		go func() {
			defer close(r.finished)
			r.result <- indexer.Serve(ctx, 20*time.Millisecond, func(context.Context) error {
				select {
				case r.calls <- struct{}{}:
				default:
				}
				return nil
			}, func() error { r.releases.Add(1); return nil })
		}()
	}
	for _, r := range runs {
		select {
		case <-r.calls:
		case <-time.After(3 * time.Second):
			t.Fatal("both independent runs did not start")
		}
	}
	runs[0].cancel()
	waitClosed(t, runs[0].finished, "first run stop")
	if err := <-runs[0].result; err != nil || runs[0].releases.Load() != 1 {
		t.Fatalf("first run = %v, releases = %d", err, runs[0].releases.Load())
	}
	// Discard observations queued before the first run joined.
	for len(runs[1].calls) != 0 {
		<-runs[1].calls
	}
	select {
	case <-runs[1].calls:
	case <-time.After(3 * time.Second):
		t.Fatal("second run stopped with the first")
	}
	if runs[1].releases.Load() != 0 {
		t.Fatal("second run released before its cancellation")
	}
	runs[1].cancel()
	waitClosed(t, runs[1].finished, "second run stop")
	if err := <-runs[1].result; err != nil || runs[1].releases.Load() != 1 {
		t.Fatalf("second run = %v, releases = %d", err, runs[1].releases.Load())
	}
}
