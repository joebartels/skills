package indexer_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"example.com/indexer"
)

var _ func(context.Context, time.Duration, func(context.Context) error, func() error) error = indexer.Serve

func TestServeInvalidIntervalHasNoEffects(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Second} {
		calls, releases := 0, 0
		err := indexer.Serve(context.Background(), interval, func(context.Context) error { calls++; return nil }, func() error { releases++; return nil })
		if !errors.Is(err, indexer.ErrInvalidInterval) || calls != 0 || releases != 0 {
			t.Fatalf("Serve = %v, calls %d, releases %d", err, calls, releases)
		}
	}
}

func TestServeAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, releaseErr := range []error{nil, errors.New("release failed")} {
		calls, releases := 0, 0
		err := indexer.Serve(ctx, time.Millisecond, func(context.Context) error { calls++; return nil }, func() error { releases++; return releaseErr })
		if calls != 0 || releases != 1 {
			t.Fatalf("calls %d, releases %d", calls, releases)
		}
		if releaseErr == nil && err != nil {
			t.Fatalf("Serve = %v", err)
		}
		if releaseErr != nil && !errors.Is(err, releaseErr) {
			t.Fatalf("release identity lost: %v", err)
		}
	}
}

func TestServeIndependentErrorCombinations(t *testing.T) {
	workErr, releaseErr := errors.New("work failed"), errors.New("release failed")
	for _, tc := range []struct {
		name          string
		work, release error
	}{
		{"work only", workErr, nil},
		{"release only", nil, releaseErr},
		{"both", workErr, releaseErr},
		{"success", nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls, releases := 0, 0
			err := indexer.Serve(ctx, time.Hour, func(got context.Context) error {
				calls++
				if got != ctx {
					t.Error("caller context changed")
				}
				if tc.work == nil {
					cancel()
				}
				return tc.work
			}, func() error { releases++; return tc.release })
			if calls != 1 || releases != 1 {
				t.Fatalf("calls %d, releases %d", calls, releases)
			}
			if tc.work == nil && tc.release == nil && err != nil {
				t.Fatalf("Serve = %v", err)
			}
			for _, want := range []error{tc.work, tc.release} {
				if want != nil && !errors.Is(err, want) {
					t.Errorf("Serve = %v, want cause %v", err, want)
				}
			}
		})
	}
}

type nonComparableError struct{ details []string }

func (e nonComparableError) Error() string { return fmt.Sprint(e.details) }

func TestServeCancellationErrorClassification(t *testing.T) {
	independent := errors.New("cleanup failed")
	for _, tc := range []struct {
		name          string
		work          error
		cause         error
		nonComparable bool
	}{
		{"nil", nil, nil, false},
		{"canceled", context.Canceled, nil, false},
		{"wrapped cancellation", fmt.Errorf("request: %w", context.Canceled), nil, false},
		{"joined ordinary cancellation", errors.Join(context.Canceled, fmt.Errorf("again: %w", context.Canceled)), nil, false},
		{"independent failure", independent, independent, false},
		{"cancellation plus failure", errors.Join(context.Canceled, independent), independent, false},
		{"non-comparable failure", nonComparableError{[]string{"failure"}}, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			releases := 0
			err := indexer.Serve(ctx, time.Hour, func(context.Context) error { cancel(); return tc.work }, func() error { releases++; return nil })
			if releases != 1 {
				t.Fatalf("releases %d", releases)
			}
			if tc.cause != nil && !errors.Is(err, tc.cause) {
				t.Fatalf("independent cause lost: %v", err)
			}
			if tc.nonComparable {
				var got nonComparableError
				if !errors.As(err, &got) {
					t.Fatalf("non-comparable error lost: %v", err)
				}
				return
			}
			if tc.cause == nil && err != nil {
				t.Fatalf("ordinary cancellation = %v", err)
			}
		})
	}
}

func TestServeJoinsCancellationCleanupBeforeRelease(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started, cleaning, allowCleanup, released := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	var unblock sync.Once
	result, done := startServe(ctx, time.Millisecond, func(got context.Context) error {
		if got != ctx {
			return errors.New("changed caller context")
		}
		close(started)
		<-got.Done()
		close(cleaning)
		<-allowCleanup
		return got.Err()
	}, func() error { close(released); return nil })
	t.Cleanup(func() { cancel(); unblock.Do(func() { close(allowCleanup) }); await(t, done) })
	await(t, started)
	cancel()
	await(t, cleaning)
	select {
	case <-released:
		t.Fatal("release overtook callback cleanup")
	case <-done:
		t.Fatal("Serve returned before cleanup")
	case <-time.After(30 * time.Millisecond):
	}
	unblock.Do(func() { close(allowCleanup) })
	await(t, done)
	await(t, released)
	if err := <-result; err != nil {
		t.Fatalf("Serve = %v", err)
	}
}

func TestServeRecursAfterCompletionWithoutOverlap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const interval = 45 * time.Millisecond
	started, finishFirst := make(chan struct{}), make(chan struct{})
	second := make(chan time.Time, 1)
	var unblock sync.Once
	calls := 0
	result, done := startServe(ctx, interval, func(context.Context) error {
		calls++
		if calls == 1 {
			close(started)
			<-finishFirst
			return nil
		}
		second <- time.Now()
		return nil
	}, func() error { return nil })
	t.Cleanup(func() { cancel(); unblock.Do(func() { close(finishFirst) }); await(t, done) })
	await(t, started)
	select {
	case <-second:
		t.Fatal("callbacks overlapped")
	case <-done:
		t.Fatal("Serve stopped early")
	case <-time.After(2 * interval):
	}
	completed := time.Now()
	unblock.Do(func() { close(finishFirst) })
	select {
	case at := <-second:
		if at.Sub(completed) < interval {
			t.Fatalf("next callback after %v; want at least %v", at.Sub(completed), interval)
		}
	case <-done:
		t.Fatal("Serve stopped before recurrence")
	case <-time.After(2 * time.Second):
		t.Fatal("no successful recurrence")
	}
	cancel()
	await(t, done)
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestServeIndependentInvocations(t *testing.T) {
	ctxA, cancelA := context.WithCancel(context.Background())
	ctxB, cancelB := context.WithCancel(context.Background())
	a, b := make(chan struct{}, 10), make(chan struct{}, 10)
	resultA, doneA := startServe(ctxA, 15*time.Millisecond, func(context.Context) error { a <- struct{}{}; return nil }, func() error { return nil })
	resultB, doneB := startServe(ctxB, 15*time.Millisecond, func(context.Context) error { b <- struct{}{}; return nil }, func() error { return nil })
	t.Cleanup(func() { cancelA(); cancelB(); await(t, doneA); await(t, doneB) })
	await(t, a)
	await(t, b)
	cancelA()
	await(t, doneA)
	if err := <-resultA; err != nil {
		t.Fatal(err)
	}
	// Drain any observation made before A joined, then require fresh work from B.
	for len(b) > 0 {
		<-b
	}
	await(t, b)
	cancelB()
	await(t, doneB)
	if err := <-resultB; err != nil {
		t.Fatal(err)
	}
}

func startServe(ctx context.Context, interval time.Duration, refresh func(context.Context) error, release func() error) (<-chan error, <-chan struct{}) {
	result, done := make(chan error, 1), make(chan struct{})
	go func() { result <- indexer.Serve(ctx, interval, refresh, release); close(done) }()
	return result, done
}

func await(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for lifecycle event")
	}
}
