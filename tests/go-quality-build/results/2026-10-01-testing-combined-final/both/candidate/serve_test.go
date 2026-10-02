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

func TestServeInvalidInterval(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Second} {
		t.Run(interval.String(), func(t *testing.T) {
			t.Parallel()
			callbacks, releases := 0, 0
			err := indexer.Serve(context.Background(), interval, func(context.Context) error { callbacks++; return nil }, func() error { releases++; return nil })
			if !errors.Is(err, indexer.ErrInvalidInterval) || callbacks != 0 || releases != 0 {
				t.Fatalf("Serve = %v, callbacks=%d releases=%d", err, callbacks, releases)
			}
		})
	}
}

func TestServeAlreadyCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls, releases := 0, 0
	err := indexer.Serve(ctx, time.Second, func(context.Context) error { calls++; return nil }, func() error { releases++; return nil })
	if err != nil || calls != 0 || releases != 1 {
		t.Fatalf("Serve = %v, callbacks=%d releases=%d", err, calls, releases)
	}
}

func TestServeErrorCombinations(t *testing.T) {
	workErr, releaseErr := errors.New("refresh failed"), errors.New("release failed")
	cases := []struct {
		name          string
		work, release error
	}{
		{"success", nil, nil},
		{"work only", workErr, nil},
		{"release only", nil, releaseErr},
		{"both", workErr, releaseErr},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			calls, releases := 0, 0
			err := indexer.Serve(ctx, time.Millisecond, func(context.Context) error {
				calls++
				if tc.work == nil {
					cancel()
				}
				return tc.work
			}, func() error { releases++; return tc.release })
			if tc.work == nil && tc.release == nil && err != nil {
				t.Errorf("Serve = %v; want nil", err)
			}
			for _, want := range []error{tc.work, tc.release} {
				if want != nil && !errors.Is(err, want) {
					t.Errorf("Serve = %v; missing cause %v", err, want)
				}
			}
			if calls != 1 || releases != 1 {
				t.Errorf("callbacks=%d releases=%d; want 1 each", calls, releases)
			}
		})
	}
}

type sliceFailure []string

func (s sliceFailure) Error() string { return fmt.Sprint([]string(s)) }

func TestServeCancellationJoinsCleanup(t *testing.T) {
	independent := sliceFailure{"independent failure"}
	cases := []struct {
		name        string
		result      error
		wantFailure bool
	}{
		{"nil cleanup result", nil, false},
		{"plain cancellation", context.Canceled, false},
		{"wrapped cancellation", fmt.Errorf("callback canceled: %w", context.Canceled), false},
		{"noncomparable independent error", independent, true},
		{"joined cancellation and independent error", errors.Join(context.Canceled, independent), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			started := make(chan context.Context, 1)
			cleaning, cleaned, gate := make(chan struct{}), make(chan struct{}), make(chan struct{})
			released, exited := make(chan struct{}), make(chan struct{})
			done := make(chan error, 1)
			var once sync.Once
			var overtook atomic.Bool
			releaseCalls := 0
			t.Cleanup(func() {
				cancel()
				once.Do(func() { close(gate) })
				select {
				case <-exited:
				case <-time.After(5 * time.Second):
					t.Error("Serve did not finish during cleanup")
				}
			})
			go func() {
				done <- indexer.Serve(ctx, time.Millisecond, func(got context.Context) error {
					started <- got
					<-got.Done()
					close(cleaning)
					<-gate
					close(cleaned)
					return tc.result
				}, func() error {
					releaseCalls++
					select {
					case <-cleaned:
					default:
						overtook.Store(true)
					}
					close(released)
					return nil
				})
				close(exited)
			}()
			select {
			case got := <-started:
				if got != ctx {
					t.Fatal("callback did not receive caller context")
				}
			case <-time.After(5 * time.Second):
				t.Fatal("callback did not start")
			}
			cancel()
			await(t, cleaning, "callback cancellation cleanup")
			// The callback has acknowledged cancellation and is held in cleanup.
			// Give an incorrectly asynchronous release a bounded chance to become visible.
			select {
			case <-released:
				t.Fatal("release overtook blocked callback cleanup")
			case <-exited:
				t.Fatal("Serve returned before callback cleanup")
			case <-time.After(30 * time.Millisecond):
			}
			once.Do(func() { close(gate) })
			select {
			case err := <-done:
				if tc.wantFailure {
					var got sliceFailure
					if !errors.As(err, &got) || len(got) != 1 || got[0] != "independent failure" {
						t.Errorf("independent error lost: %v", err)
					}
				} else if err != nil {
					t.Errorf("ordinary cancellation = %v; want nil", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Serve did not join callback")
			}
			await(t, exited, "Serve return")
			if overtook.Load() || releaseCalls != 1 {
				t.Errorf("release calls=%d, overtook=%v", releaseCalls, overtook.Load())
			}
		})
	}
}

func TestServeRecursAfterCompletionWithoutOverlap(t *testing.T) {
	t.Parallel()
	const interval = 40 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	started, gate := make(chan struct{}), make(chan struct{})
	completed, second := make(chan time.Time, 1), make(chan time.Time, 1)
	done, exited := make(chan error, 1), make(chan struct{})
	var calls, active, releases atomic.Int32
	var overlap atomic.Bool
	var once sync.Once
	t.Cleanup(func() {
		cancel()
		once.Do(func() { close(gate) })
		select {
		case <-exited:
		case <-time.After(5 * time.Second):
			t.Error("recurring Serve did not stop during cleanup")
		}
	})
	go func() {
		done <- indexer.Serve(ctx, interval, func(context.Context) error {
			if active.Add(1) != 1 {
				overlap.Store(true)
			}
			defer active.Add(-1)
			switch calls.Add(1) {
			case 1:
				close(started)
				select {
				case <-gate:
				case <-ctx.Done():
					return ctx.Err()
				}
				// Report the return event, rather than treating gate release as completion.
				defer func() { completed <- time.Now() }()
			case 2:
				second <- time.Now()
			}
			return nil
		}, func() error { releases.Add(1); return nil })
		close(exited)
	}()
	await(t, started, "immediate refresh")
	// This deliberate hold exceeds the interval; a ticker started before callback
	// completion would already be ready when the first callback returns.
	timer := time.NewTimer(2 * interval)
	select {
	case <-timer.C:
	case <-exited:
		timer.Stop()
		t.Fatal("Serve returned while first callback was blocked")
	}
	once.Do(func() { close(gate) })
	var completion time.Time
	select {
	case completion = <-completed:
	case <-time.After(5 * time.Second):
		t.Fatal("first callback did not complete")
	}
	select {
	case next := <-second:
		if elapsed := next.Sub(completion); elapsed < interval-time.Millisecond {
			t.Errorf("next refresh after %s; want at least %s from completion", elapsed, interval)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("successful refresh did not recur")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Serve = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not stop")
	}
	await(t, exited, "recurring Serve return")
	if overlap.Load() || active.Load() != 0 || releases.Load() != 1 {
		t.Errorf("overlap=%v active=%d releases=%d", overlap.Load(), active.Load(), releases.Load())
	}
}

func TestServeIndependentInvocations(t *testing.T) {
	t.Parallel()
	type run struct {
		cancel          context.CancelFunc
		events          chan int32
		done            chan error
		exited          chan struct{}
		calls, releases atomic.Int32
	}
	start := func() *run {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		r := &run{cancel: cancel, events: make(chan int32, 100), done: make(chan error, 1), exited: make(chan struct{})}
		t.Cleanup(func() {
			cancel()
			select {
			case <-r.exited:
			case <-time.After(5 * time.Second):
				t.Error("independent Serve failed to stop")
			}
		})
		go func() {
			r.done <- indexer.Serve(ctx, 10*time.Millisecond, func(ctx context.Context) error {
				count := r.calls.Add(1)
				select {
				case r.events <- count:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}, func() error { r.releases.Add(1); return nil })
			close(r.exited)
		}()
		return r
	}
	first, second := start(), start()
	for _, r := range []*run{first, second} {
		select {
		case <-r.events:
		case <-time.After(5 * time.Second):
			t.Fatal("independent run did not start")
		}
	}
	first.cancel()
	await(t, first.exited, "first independent run stop")
	if err := <-first.done; err != nil {
		t.Errorf("first Serve = %v", err)
	}
	baseline := second.calls.Load()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for {
		select {
		case count := <-second.events:
			if count <= baseline {
				continue
			}
			if second.releases.Load() != 0 {
				t.Fatal("stopping first released second")
			}
			second.cancel()
			await(t, second.exited, "second independent run stop")
			if err := <-second.done; err != nil {
				t.Errorf("second Serve = %v", err)
			}
			if first.releases.Load() != 1 || second.releases.Load() != 1 {
				t.Error("each run must release once")
			}
			return
		case <-timer.C:
			t.Fatal("second run did not recur after first stopped")
		}
	}
}
