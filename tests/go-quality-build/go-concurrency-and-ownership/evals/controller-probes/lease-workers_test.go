package leaseworkers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type recoveryLease struct {
	run   func(context.Context, Job) error
	close func() error
}

func (l recoveryLease) Run(ctx context.Context, job Job) error { return l.run(ctx, job) }
func (l recoveryLease) Close() error                           { return l.close() }

func recoveryAwait(t *testing.T, event <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-event:
	case <-time.After(time.Second):
		t.Fatalf("missing %s", what)
	}
}

func recoveryHeld(t *testing.T, done <-chan error, what string) {
	t.Helper()
	select {
	case err := <-done:
		t.Fatalf("returned before %s: %v", what, err)
	case <-time.After(20 * time.Millisecond):
	}
}

func TestRecoverySparseProgressAndReleaseCapacity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs := make(chan Job, 3)
	started := make(chan Job, 3)
	closeOne := make(chan struct{})
	runOne := make(chan struct{})
	runTwo := make(chan struct{})
	releaseOne := make(chan struct{})
	var release sync.Once
	defer release.Do(func() { close(releaseOne) })
	var live, maxLive atomic.Int32
	var earlyClose atomic.Bool
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, jobs, 2, func(ctx context.Context, job Job) (Lease, error) {
			n := live.Add(1)
			for old := maxLive.Load(); n > old && !maxLive.CompareAndSwap(old, n); old = maxLive.Load() {
			}
			var running atomic.Bool
			return recoveryLease{
				run: func(ctx context.Context, job Job) error {
					running.Store(true)
					defer running.Store(false)
					started <- job
					var gate <-chan struct{}
					if job == 1 {
						gate = runOne
					}
					if job == 2 {
						gate = runTwo
					}
					if gate != nil {
						select {
						case <-gate:
						case <-ctx.Done():
							return ctx.Err()
						}
					}
					return nil
				},
				close: func() error {
					if running.Load() {
						earlyClose.Store(true)
					}
					if job == 1 {
						close(closeOne)
						<-releaseOne
					}
					live.Add(-1)
					return nil
				},
			}, nil
		})
	}()
	waitJob := func(want Job) {
		t.Helper()
		select {
		case got := <-started:
			if got != want {
				t.Fatalf("started %d, want %d", got, want)
			}
		case <-time.After(time.Second):
			t.Fatalf("job %d did not start with spare capacity and open input", want)
		}
	}
	jobs <- 1
	waitJob(1)
	jobs <- 2
	waitJob(2)
	close(runOne)
	recoveryAwait(t, closeOne, "first Close")
	jobs <- 3
	select {
	case got := <-started:
		t.Fatalf("job %d started before Close released its slot", got)
	case <-time.After(20 * time.Millisecond):
	}
	release.Do(func() { close(releaseOne) })
	waitJob(3) // Unrelated Run 2 remains held; no whole-cohort barrier.
	close(runTwo)
	close(jobs)
	select {
	case err := <-done:
		if err != nil || live.Load() != 0 || maxLive.Load() > 2 || earlyClose.Load() {
			t.Fatalf("err=%v live=%d max=%d earlyClose=%v", err, live.Load(), maxLive.Load(), earlyClose.Load())
		}
	case <-time.After(time.Second):
		t.Fatal("Serve did not join completed work")
	}
}

func TestRecoveryFailureSupervisesBlockedAcquisition(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2 // Deliberately stays open after these jobs.
	secondOpen := make(chan struct{})
	secondStopped := make(chan struct{})
	releaseOpen := make(chan struct{})
	closeStarted := make(chan struct{})
	releaseClose := make(chan struct{})
	var openOnce, closeOnce sync.Once
	defer openOnce.Do(func() { close(releaseOpen) })
	defer closeOnce.Do(func() { close(releaseClose) })
	independent := errors.New("independent run failure")
	closeErr := errors.New("late lease close failure")
	var lateRuns, closes atomic.Int32
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, jobs, 2, func(ctx context.Context, job Job) (Lease, error) {
			if job == 1 {
				return recoveryLease{run: func(context.Context, Job) error {
					select {
					case <-secondOpen:
					case <-ctx.Done():
						return ctx.Err()
					}
					return errors.Join(context.DeadlineExceeded, independent)
				}, close: func() error { closes.Add(1); <-releaseClose; return nil }}, nil
			}
			close(secondOpen)
			<-ctx.Done()
			close(secondStopped)
			<-releaseOpen
			return recoveryLease{run: func(context.Context, Job) error {
				lateRuns.Add(1)
				return nil
			}, close: func() error {
				closes.Add(1)
				close(closeStarted)
				<-releaseClose
				return closeErr
			}}, nil
		})
	}()
	recoveryAwait(t, secondStopped, "stop while acquisition blocks")
	recoveryHeld(t, done, "late acquisition returns")
	openOnce.Do(func() { close(releaseOpen) })
	recoveryAwait(t, closeStarted, "late acquired lease closure")
	recoveryHeld(t, done, "held Close completion")
	closeOnce.Do(func() { close(releaseClose) })
	select {
	case err := <-done:
		if !errors.Is(err, independent) || !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, closeErr) || lateRuns.Load() != 0 || closes.Load() != 2 {
			t.Fatalf("err=%v lateRuns=%d closes=%d", err, lateRuns.Load(), closes.Load())
		}
	case <-time.After(time.Second):
		t.Fatal("Serve did not finish after all owned work completed")
	}
	close(jobs) // A borrowed channel must still be open.
}

func TestRecoveryCloseFailureAndCanceledEntry(t *testing.T) {
	closeErr := errors.New("release failed")
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	opened := 0
	err := Serve(context.Background(), jobs, 1, func(context.Context, Job) (Lease, error) {
		opened++
		return recoveryLease{run: func(context.Context, Job) error { return nil }, close: func() error { return closeErr }}, nil
	})
	if opened != 1 || !errors.Is(err, closeErr) {
		t.Fatalf("opened=%d err=%v", opened, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = Serve(ctx, jobs, 1, func(context.Context, Job) (Lease, error) {
		t.Error("admitted canceled work")
		return nil, errors.New("unexpected Open")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled entry err=%v", err)
	}
}
