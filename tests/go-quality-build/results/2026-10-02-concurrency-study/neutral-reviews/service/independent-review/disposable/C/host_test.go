package workers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var _ func(context.Context, <-chan Job, int, func(context.Context, Job) (Lease, error)) error = Serve

type funcLease struct {
	run   func(context.Context, Job) error
	close func() error
}

func (l *funcLease) Run(ctx context.Context, job Job) error { return l.run(ctx, job) }
func (l *funcLease) Close() error                           { return l.close() }

func await[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for Serve event")
		var zero T
		return zero
	}
}

type basicLease struct {
	accepted *atomic.Int64
	closed   *atomic.Int64
}

func (l *basicLease) Run(context.Context, Job) error { l.accepted.Add(1); return nil }
func (l *basicLease) Close() error                   { l.closed.Add(1); return nil }
func TestServeFiniteJobs(t *testing.T) {
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	var accepted, closed atomic.Int64
	err := Serve(context.Background(), jobs, 2, func(context.Context, Job) (Lease, error) { return &basicLease{&accepted, &closed}, nil })
	if err != nil || accepted.Load() != 2 || closed.Load() != 2 {
		t.Fatalf("accepted=%d closed=%d err=%v", accepted.Load(), closed.Load(), err)
	}
}

func TestServeCapacityAndRelease(t *testing.T) {
	jobs := make(chan Job, 3)
	for _, job := range []Job{1, 2, 3} {
		jobs <- job
	}
	close(jobs)

	started := make(chan Job, 3)
	runGate := make(chan struct{})
	closeEntered := make(chan struct{}, 1)
	closeGate := make(chan struct{})
	var releaseRuns, releaseClose sync.Once
	var live, maxLive, closed atomic.Int64
	badAdmission := make(chan Job, 1)
	result := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		result <- Serve(context.Background(), jobs, 2, func(_ context.Context, job Job) (Lease, error) {
			if job == 3 && closed.Load() != 2 {
				badAdmission <- job
			}
			n := live.Add(1)
			for {
				old := maxLive.Load()
				if n <= old || maxLive.CompareAndSwap(old, n) {
					break
				}
			}
			return &funcLease{
				run: func(context.Context, Job) error {
					started <- job
					<-runGate
					return nil
				},
				close: func() error {
					if job == 1 {
						closeEntered <- struct{}{}
						<-closeGate
					}
					live.Add(-1)
					closed.Add(1)
					return nil
				},
			}, nil
		})
	}()
	t.Cleanup(func() {
		releaseRuns.Do(func() { close(runGate) })
		releaseClose.Do(func() { close(closeGate) })
		await(t, done)
	})
	first, second := await(t, started), await(t, started)
	if first == second || first == 3 || second == 3 {
		t.Fatalf("first started jobs = %d, %d; want 1 and 2", first, second)
	}
	releaseRuns.Do(func() { close(runGate) })
	await(t, closeEntered)
	if live.Load() != 2 {
		t.Fatalf("live leases during blocked Close = %d, want 2", live.Load())
	}
	releaseClose.Do(func() { close(closeGate) })
	if err := await(t, result); err != nil {
		t.Fatalf("Serve: %v", err)
	}
	select {
	case job := <-badAdmission:
		t.Fatalf("job %d admitted before earlier closes completed", job)
	default:
	}
	if got := await(t, started); got != 3 {
		t.Fatalf("last started job = %d, want 3", got)
	}
	if maxLive.Load() != 2 || live.Load() != 0 || closed.Load() != 3 {
		t.Fatalf("maxLive=%d live=%d closed=%d", maxLive.Load(), live.Load(), closed.Load())
	}
}

func TestServeFailureJoinsBeforeCloseAndRetainsErrors(t *testing.T) {
	runErr := errors.New("run failed")
	closeErr := errors.New("close failed")
	jobs := make(chan Job, 3)
	jobs <- 1
	jobs <- 2
	jobs <- 3
	close(jobs)
	secondStarted := make(chan struct{})
	cleanupStarted := make(chan struct{})
	cleanupGate := make(chan struct{})
	var release sync.Once
	var runTwoDone atomic.Bool
	var opened, closed atomic.Int64
	badClose := make(chan struct{}, 1)
	result := make(chan error, 1)
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(done)
		result <- Serve(ctx, jobs, 2, func(_ context.Context, job Job) (Lease, error) {
			opened.Add(1)
			return &funcLease{
				run: func(ctx context.Context, _ Job) error {
					if job == 1 {
						<-secondStarted
						return runErr
					}
					close(secondStarted)
					<-ctx.Done()
					close(cleanupStarted)
					<-cleanupGate
					runTwoDone.Store(true)
					return nil
				},
				close: func() error {
					if !runTwoDone.Load() {
						badClose <- struct{}{}
					}
					closed.Add(1)
					if job == 1 {
						return closeErr
					}
					return nil
				},
			}, nil
		})
	}()
	t.Cleanup(func() {
		cancel()
		release.Do(func() { close(cleanupGate) })
		await(t, done)
	})
	await(t, cleanupStarted)
	cancel() // Cancellation observed during cleanup must also remain inspectable.
	release.Do(func() { close(cleanupGate) })
	err := await(t, result)
	for _, want := range []error{runErr, closeErr, context.Canceled} {
		if !errors.Is(err, want) {
			t.Errorf("Serve error %v does not retain %v", err, want)
		}
	}
	if opened.Load() != 2 || closed.Load() != 2 {
		t.Errorf("opened=%d closed=%d, want 2 each", opened.Load(), closed.Load())
	}
	select {
	case <-badClose:
		t.Error("Close ran before all Runs completed")
	default:
	}
}

func TestServeOpenFailureStopsAndClosesAcquiredLease(t *testing.T) {
	openErr := errors.New("open failed")
	closeErr := errors.New("close failed")
	jobs := make(chan Job, 3)
	jobs <- 1
	jobs <- 2
	jobs <- 3
	close(jobs)
	var opened, closed atomic.Int64
	var runDone atomic.Bool
	started := make(chan struct{})
	err := Serve(context.Background(), jobs, 2, func(_ context.Context, job Job) (Lease, error) {
		opened.Add(1)
		if job == 2 {
			<-started
			return nil, openErr
		}
		return &funcLease{
			run: func(ctx context.Context, _ Job) error {
				close(started)
				<-ctx.Done()
				runDone.Store(true)
				return nil
			},
			close: func() error {
				if !runDone.Load() {
					t.Error("Close preceded Run completion")
				}
				closed.Add(1)
				return closeErr
			},
		}, nil
	})
	if !errors.Is(err, openErr) || !errors.Is(err, closeErr) || opened.Load() != 2 || closed.Load() != 1 {
		t.Fatalf("err=%v opened=%d closed=%d", err, opened.Load(), closed.Load())
	}
}

func TestServeAcquiredLeaseWithoutRunIsClosed(t *testing.T) {
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	ctx, cancel := context.WithCancel(context.Background())
	firstStarted := make(chan struct{})
	secondOpening := make(chan struct{})
	cleanupStarted := make(chan struct{})
	cleanupGate := make(chan struct{})
	var release sync.Once
	var firstDone, secondRan atomic.Bool
	var closed atomic.Int64
	result := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		result <- Serve(ctx, jobs, 2, func(openCtx context.Context, job Job) (Lease, error) {
			if job == 2 {
				close(secondOpening)
				<-openCtx.Done()
				return &funcLease{
					run:   func(context.Context, Job) error { secondRan.Store(true); return nil },
					close: func() error { closed.Add(1); return nil },
				}, nil
			}
			return &funcLease{
				run: func(runCtx context.Context, _ Job) error {
					close(firstStarted)
					<-runCtx.Done()
					close(cleanupStarted)
					<-cleanupGate
					firstDone.Store(true)
					return nil
				},
				close: func() error {
					if !firstDone.Load() {
						return errors.New("first lease closed before Run completed")
					}
					closed.Add(1)
					return nil
				},
			}, nil
		})
	}()
	t.Cleanup(func() {
		cancel()
		release.Do(func() { close(cleanupGate) })
		await(t, done)
	})
	await(t, firstStarted)
	await(t, secondOpening)
	cancel()
	await(t, cleanupStarted)
	release.Do(func() { close(cleanupGate) })
	if err := await(t, result); !errors.Is(err, context.Canceled) {
		t.Fatalf("Serve = %v, want cancellation", err)
	}
	if secondRan.Load() || closed.Load() != 2 {
		t.Fatalf("secondRan=%t closed=%d, want false and 2", secondRan.Load(), closed.Load())
	}
}

func TestServeIndependentRunAndCloseErrors(t *testing.T) {
	runErr := errors.New("run failed")
	closeErr := errors.New("close failed")
	for _, tc := range []struct {
		name             string
		runErr, closeErr error
	}{
		{"run only", runErr, nil},
		{"close only", nil, closeErr},
		{"both", runErr, closeErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			jobs := make(chan Job, 2)
			jobs <- 1
			jobs <- 2
			close(jobs)
			closes, opens := 0, 0
			err := Serve(context.Background(), jobs, 1, func(context.Context, Job) (Lease, error) {
				opens++
				return &funcLease{
					run:   func(context.Context, Job) error { return tc.runErr },
					close: func() error { closes++; return tc.closeErr },
				}, nil
			})
			for _, want := range []error{tc.runErr, tc.closeErr} {
				if want != nil && !errors.Is(err, want) {
					t.Errorf("Serve error %v does not retain %v", err, want)
				}
			}
			if opens != 1 || closes != 1 {
				t.Errorf("Open calls = %d, Close calls = %d; want 1 each", opens, closes)
			}
		})
	}
}

func TestServeKeepsCompletedEffectsBeforeLaterFailure(t *testing.T) {
	runErr := errors.New("later run failed")
	jobs := make(chan Job, 3)
	for _, job := range []Job{1, 2, 3} {
		jobs <- job
	}
	close(jobs)
	var effects []Job
	var opened, closed int
	err := Serve(context.Background(), jobs, 1, func(_ context.Context, job Job) (Lease, error) {
		opened++
		return &funcLease{
			run: func(context.Context, Job) error {
				if job == 2 {
					return runErr
				}
				effects = append(effects, job)
				return nil
			},
			close: func() error { closed++; return nil },
		}, nil
	})
	if !errors.Is(err, runErr) || opened != 2 || closed != 2 || len(effects) != 1 || effects[0] != 1 {
		t.Fatalf("err=%v opened=%d closed=%d effects=%v", err, opened, closed, effects)
	}
}

func TestServeCancellationWhileWaitingForInput(t *testing.T) {
	jobs := make(chan Job, 1)
	jobs <- 1
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	var closed atomic.Int64
	result := make(chan error, 1)
	go func() {
		result <- Serve(ctx, jobs, 2, func(context.Context, Job) (Lease, error) {
			return &funcLease{
				run: func(runCtx context.Context, _ Job) error {
					close(started)
					<-runCtx.Done()
					return nil
				},
				close: func() error { closed.Add(1); return nil },
			}, nil
		})
	}()
	t.Cleanup(cancel)
	await(t, started)
	cancel()
	if err := await(t, result); !errors.Is(err, context.Canceled) {
		t.Fatalf("Serve = %v, want cancellation", err)
	}
	if closed.Load() != 1 {
		t.Fatalf("Close calls = %d, want 1", closed.Load())
	}
	close(jobs) // The caller still owns the input channel.
}

func TestServeAlreadyCanceledOpensNothing(t *testing.T) {
	jobs := make(chan Job, 1)
	jobs <- 1
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	opened := false
	err := Serve(ctx, jobs, 2, func(context.Context, Job) (Lease, error) {
		opened = true
		return nil, errors.New("unexpected open")
	})
	if opened || !errors.Is(err, context.Canceled) {
		t.Fatalf("opened=%t err=%v", opened, err)
	}
}
