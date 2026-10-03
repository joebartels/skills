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
            return &testLease{run: func(ctx context.Context, _ Job) error {
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
            return &testLease{run: func(context.Context, Job) error {
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

func TestReviewStopJoinAllCausesAndCloseCompletion(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    jobs := make(chan Job, 3)
    jobs <- 1
    jobs <- 2
    jobs <- 3
    close(jobs)
    openErr, runErr, closeErr := errors.New("open"), errors.New("run"), errors.New("close")
    running, stopped := make(chan struct{}), make(chan struct{})
    cleanup, closeEntered, releaseClose := make(chan struct{}), make(chan struct{}), make(chan struct{})
    var opens, closes atomic.Int32
    done := make(chan error, 1)
    go func() {
        done <- Serve(ctx, jobs, 2, func(workCtx context.Context, j Job) (Lease, error) {
            opens.Add(1)
            if j == 2 {
                <-running
                return nil, openErr
            }
            return &testLease{run: func(workCtx context.Context, _ Job) error {
                close(running)
                <-workCtx.Done()
                close(stopped)
                <-cleanup
                return runErr
            }, close: func() error {
                closes.Add(1)
                close(closeEntered)
                <-releaseClose
                return closeErr
            }}, nil
        })
    }()
    reviewEvent(t, stopped, "Run observed coordinated stop")
    cancel()
    select {
    case <-closeEntered:
        t.Error("Close entered before Run cleanup completed")
    case <-done:
        t.Fatal("Serve returned before cleanup")
    case <-time.After(30 * time.Millisecond):
    }
    close(cleanup)
    reviewEvent(t, closeEntered, "Close after Run joined")
    select {
    case <-done:
        t.Fatal("Serve returned before Close completion")
    case <-time.After(30 * time.Millisecond):
    }
    close(releaseClose)
    err := reviewDone(t, done)
    for _, want := range []error{openErr, runErr, closeErr, context.Canceled} {
        if !errors.Is(err, want) { t.Errorf("missing %v in %v", want, err) }
    }
    if opens.Load() != 2 || closes.Load() != 1 || len(jobs) != 1 {
        t.Errorf("opens=%d closes=%d queued=%d; want 2,1,1", opens.Load(), closes.Load(), len(jobs))
    }
}

func TestReviewAcquiredLeaseWhoseRunNeverStarts(t *testing.T) {
    jobs := make(chan Job, 2)
    jobs <- 1
    jobs <- 2
    close(jobs)
    acquired := make(chan struct{})
    openErr := errors.New("open")
    var runs, closes atomic.Int32
    err := Serve(context.Background(), jobs, 2, func(workCtx context.Context, j Job) (Lease, error) {
        if j == 1 {
            <-acquired
            return nil, openErr
        }
        lease := &testLease{run: func(context.Context, Job) error { runs.Add(1); return nil },
            close: func() error { closes.Add(1); return nil }}
        close(acquired)
        <-workCtx.Done()
        return lease, nil
    })
    if !errors.Is(err, openErr) || runs.Load() != 0 || closes.Load() != 1 {
        t.Errorf("err=%v runs=%d closes=%d; want Open cause,0,1", err, runs.Load(), closes.Load())
    }
}

func TestReviewCapacityIncludesCloseCompletion(t *testing.T) {
    jobs := make(chan Job, 3)
    jobs <- 1
    jobs <- 2
    jobs <- 3
    close(jobs)
    closeEntered, releaseClose := make(chan struct{}, 3), make(chan struct{})
    var openCount, held, maxHeld atomic.Int32
    done := make(chan error, 1)
    go func() {
        done <- Serve(context.Background(), jobs, 2, func(context.Context, Job) (Lease, error) {
            openCount.Add(1)
            n := held.Add(1)
            for old := maxHeld.Load(); old < n && !maxHeld.CompareAndSwap(old, n); old = maxHeld.Load() {}
            return &testLease{run: func(context.Context, Job) error { return nil }, close: func() error {
                closeEntered <- struct{}{}
                <-releaseClose
                held.Add(-1)
                return nil
            }}, nil
        })
    }()
    reviewEvent(t, closeEntered, "held Close entered")
    if openCount.Load() != 2 {
        t.Errorf("acquired %d leases before first held Close completed; want 2", openCount.Load())
    }
    close(releaseClose)
    if err := reviewDone(t, done); err != nil { t.Fatal(err) }
    if maxHeld.Load() != 2 || held.Load() != 0 || openCount.Load() != 3 {
        t.Errorf("maxHeld=%d held=%d opens=%d; want 2,0,3", maxHeld.Load(), held.Load(), openCount.Load())
    }
}

func TestReviewRunFailureStopsSibling(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    jobs := make(chan Job, 2)
    jobs <- 1
    jobs <- 2
    close(jobs)
    started, barrier := make(chan struct{}, 2), make(chan struct{})
    runErr := errors.New("run")
    var closes atomic.Int32
    done := make(chan error, 1)
    go func() {
        done <- Serve(ctx, jobs, 2, func(context.Context, Job) (Lease, error) {
            return &testLease{run: func(workCtx context.Context, j Job) error {
                started <- struct{}{}
                <-barrier
                if j == 1 { return runErr }
                <-workCtx.Done()
                return workCtx.Err()
            }, close: func() error { closes.Add(1); return nil }}, nil
        })
    }()
    reviewEvent(t, started, "first Run")
    reviewEvent(t, started, "second Run")
    close(barrier)
    missing := false
    var err error
    select {
    case err = <-done:
    case <-time.After(250 * time.Millisecond):
        missing = true
        cancel()
        err = reviewDone(t, done)
    }
    if !errors.Is(err, runErr) || closes.Load() != 2 {
        t.Errorf("err=%v closes=%d; want Run failure and 2 closes", err, closes.Load())
    }
    if missing { t.Error("Run failure did not stop sibling Run") }
}
