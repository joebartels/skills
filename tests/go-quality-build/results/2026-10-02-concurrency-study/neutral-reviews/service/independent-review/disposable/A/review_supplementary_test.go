package workers

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func reviewEvent(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatalf("missing event: %s", what)
	}
}

func reviewDone(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(time.Second):
		t.Fatal("Serve did not join owned work")
		return nil
	}
}

func TestReviewAvailableJobWithOpenInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs := make(chan Job, 1)
	jobs <- 1
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, jobs, 2, func(context.Context, Job) (Lease, error) {
			return &supplementaryLease{run: func(ctx context.Context, _ Job) error {
				close(started)
				<-ctx.Done()
				return ctx.Err()
			}, close: func() error { return nil }}, nil
		})
	}()
	missing := false
	select {
	case <-started:
	case <-time.After(250 * time.Millisecond):
		missing = true
	}
	cancel()
	err := reviewDone(t, done)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("cancellation not retained: %v", err)
	}
	if missing {
		t.Error("available job did not start with capacity 2 and caller input left open")
	}
}

func TestReviewRunFailureCancelsPendingOpen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	runErr := errors.New("independent Run failure")
	pendingOpen, runFailed := make(chan struct{}), make(chan struct{})
	var closes atomic.Int32
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, jobs, 2, func(workCtx context.Context, j Job) (Lease, error) {
			if j == 2 {
				close(pendingOpen)
				<-workCtx.Done()
				return nil, workCtx.Err()
			}
			return &supplementaryLease{run: func(context.Context, Job) error {
				<-pendingOpen
				close(runFailed)
				return runErr
			}, close: func() error { closes.Add(1); return nil }}, nil
		})
	}()
	reviewEvent(t, runFailed, "Run 1 failed while Open 2 awaited stop")
	failed := false
	var err error
	select {
	case err = <-done:
	case <-time.After(250 * time.Millisecond):
		failed = true
		cancel()
		err = reviewDone(t, done)
	}
	if !errors.Is(err, runErr) || closes.Load() != 1 {
		t.Errorf("err=%v closes=%d; want Run cause and one owned Close", err, closes.Load())
	}
	if failed {
		t.Error("Run failure did not cancel cooperating pending Open; caller cancellation was required")
	}
}

type supplementaryLease struct {
	run   func(context.Context, Job) error
	close func() error
}

func (l *supplementaryLease) Run(ctx context.Context, j Job) error { return l.run(ctx, j) }
func (l *supplementaryLease) Close() error                         { return l.close() }
