package workers

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type testLease struct {
	run   func(context.Context, Job) error
	close func() error
}

func (l *testLease) Run(ctx context.Context, j Job) error { return l.run(ctx, j) }
func (l *testLease) Close() error                         { return l.close() }

func waitSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}
func waitJob(t *testing.T, ch <-chan Job) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for job start")
	}
}

func TestServeCapacityAndLifecycleBoundary(t *testing.T) {
	jobs := make(chan Job, 3)
	jobs <- 1
	jobs <- 2
	jobs <- 3
	close(jobs)
	started := make(chan Job, 3)
	releaseRuns := make(chan struct{})
	closeEntered := make(chan struct{}, 3)
	releaseClose := make(chan struct{})
	var mu sync.Mutex
	running, peak, closes := 0, 0, 0
	done := make(chan error, 1)
	go func() {
		done <- Serve(context.Background(), jobs, 2, func(_ context.Context, j Job) (Lease, error) {
			return &testLease{run: func(context.Context, Job) error {
				mu.Lock()
				running++
				if running > peak {
					peak = running
				}
				mu.Unlock()
				started <- j
				<-releaseRuns
				mu.Lock()
				running--
				mu.Unlock()
				return nil
			}, close: func() error {
				closeEntered <- struct{}{}
				<-releaseClose
				mu.Lock()
				closes++
				mu.Unlock()
				return nil
			}}, nil
		})
	}()
	waitJob(t, started)
	waitJob(t, started)
	mu.Lock()
	gotPeak := peak
	mu.Unlock()
	if gotPeak != 2 {
		t.Fatalf("peak concurrent Runs=%d, want 2", gotPeak)
	}
	select {
	case <-started:
		t.Fatal("admitted beyond capacity")
	default:
	}
	close(releaseRuns)
	// Both cohort Runs have completed before any Close begins.
	waitSignal(t, closeEntered)
	select {
	case err := <-done:
		t.Fatalf("Serve returned before Close completed: %v", err)
	default:
	}
	close(releaseClose)
	waitSignal(t, closeEntered)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if closes != 3 {
		t.Fatalf("closed=%d, want 3", closes)
	}
}

func TestServeFailureJoinsAndClosesOwnedLeases(t *testing.T) {
	runErr, closeErr := errors.New("run"), errors.New("close")
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	var closed int
	err := Serve(context.Background(), jobs, 2, func(_ context.Context, j Job) (Lease, error) {
		return &testLease{run: func(ctx context.Context, _ Job) error {
			if j == 1 {
				return runErr
			}
			<-ctx.Done()
			return nil
		}, close: func() error {
			closed++
			if j == 1 {
				return closeErr
			}
			return nil
		}}, nil
	})
	if !errors.Is(err, runErr) {
		t.Fatalf("missing run error: %v", err)
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("missing close error: %v", err)
	}
	if closed != 2 {
		t.Fatalf("closed=%d, want both acquired leases", closed)
	}
}

func TestServeRetainsOpenFailure(t *testing.T) {
	openErr := errors.New("open")
	jobs := make(chan Job, 1)
	jobs <- 1
	close(jobs)
	err := Serve(context.Background(), jobs, 1, func(context.Context, Job) (Lease, error) { return nil, openErr })
	if !errors.Is(err, openErr) {
		t.Fatalf("error=%v, want wrapped open error", err)
	}
}

func TestServeCancellationInterruptsOpenAndClosesReturnedLease(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan Job) // deliberately left open
	var closed int
	openStarted := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, jobs, 1, func(ctx context.Context, _ Job) (Lease, error) {
			close(openStarted)
			<-ctx.Done()
			return &testLease{run: func(context.Context, Job) error { return nil }, close: func() error { closed++; return nil }}, nil
		})
	}()
	// Send one job, then cancellation interrupts cooperative Open.
	// A receiver is observable through the open callback's context wait only after this send.
	// Use an unbuffered send, so Serve has selected the input before cancellation.
	// Cancellation is checked before each subsequent admission.
	//
	// Synchronize callback entry with an input sender that completes the rendezvous.
	sent := make(chan struct{})
	go func() { jobs <- 1; close(sent) }()
	waitSignal(t, sent)
	waitSignal(t, openStarted)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error=%v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not stop after cancellation")
	}
	if closed != 1 {
		t.Fatalf("closed=%d, want 1", closed)
	}
}

func TestServeAlreadyCanceledDoesNotOpen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	if err := Serve(ctx, make(chan Job), 1, func(context.Context, Job) (Lease, error) { called = true; return nil, nil }); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
	if called {
		t.Fatal("opened a lease for an already-canceled call")
	}
}
