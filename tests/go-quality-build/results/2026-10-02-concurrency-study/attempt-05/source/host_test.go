package workers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testLease struct {
	run   func(context.Context, Job) error
	close func() error
}

func (l *testLease) Run(ctx context.Context, j Job) error { return l.run(ctx, j) }
func (l *testLease) Close() error                         { return l.close() }

func waitSignal(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func TestServeRunsConcurrentlyAndJoinsCohortBeforeClose(t *testing.T) {
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	var finished atomic.Int32
	var closed atomic.Int32
	served := make(chan error, 1)
	go func() {
		served <- Serve(context.Background(), jobs, 2, func(context.Context, Job) (Lease, error) {
			return &testLease{
				run: func(context.Context, Job) error { started <- struct{}{}; <-release; finished.Add(1); return nil },
				close: func() error {
					if finished.Load() != 2 {
						t.Errorf("lease closed before cohort Runs finished: %d", finished.Load())
					}
					closed.Add(1)
					return nil
				},
			}, nil
		})
	}()
	waitSignal(t, started, "first Run")
	waitSignal(t, started, "second concurrent Run")
	select {
	case err := <-served:
		t.Fatalf("Serve returned before Runs completed: %v", err)
	default:
	}
	close(release)
	select {
	case err := <-served:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not finish after Runs were released")
	}
	if closed.Load() != 2 {
		t.Fatalf("closed=%d, want 2", closed.Load())
	}
}

func TestServeRetainsIndependentRunAndCloseErrors(t *testing.T) {
	runErr, closeErr := errors.New("run failed"), errors.New("close failed")
	jobs := make(chan Job, 1)
	jobs <- 7
	close(jobs)
	err := Serve(context.Background(), jobs, 1, func(context.Context, Job) (Lease, error) {
		return &testLease{run: func(context.Context, Job) error { return runErr }, close: func() error { return closeErr }}, nil
	})
	if !errors.Is(err, runErr) || !errors.Is(err, closeErr) {
		t.Fatalf("Serve error %v must retain Run and Close errors", err)
	}
}

func TestServeCancellationInterruptsOpenInputWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan Job)
	var opened atomic.Int32
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, jobs, 2, func(context.Context, Job) (Lease, error) { opened.Add(1); return nil, nil })
	}()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Serve error=%v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve remained blocked on open jobs channel after cancellation")
	}
	if opened.Load() != 0 {
		t.Fatalf("opened=%d after already-canceled call", opened.Load())
	}
}

func TestServeClosesAcquiredLeaseWhenOpenPeerFails(t *testing.T) {
	openErr := errors.New("open failed")
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	var once sync.Once
	var closed atomic.Int32
	err := Serve(context.Background(), jobs, 2, func(_ context.Context, job Job) (Lease, error) {
		if job == 2 {
			once.Do(func() {})
			return nil, openErr
		}
		return &testLease{run: func(context.Context, Job) error { t.Error("Run started after cohort Open failure"); return nil }, close: func() error { closed.Add(1); return nil }}, nil
	})
	if !errors.Is(err, openErr) {
		t.Fatalf("Serve error=%v, want Open error", err)
	}
	if closed.Load() != 1 {
		t.Fatalf("closed=%d, want acquired lease released once", closed.Load())
	}
}
