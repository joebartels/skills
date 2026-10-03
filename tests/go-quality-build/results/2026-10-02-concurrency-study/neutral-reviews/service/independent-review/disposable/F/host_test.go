package workers

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

const waitBound = 5 * time.Second

func await[T any](t *testing.T, ch <-chan T, what string) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(waitBound):
		t.Fatalf("timed out waiting for %s", what)
		var zero T
		return zero
	}
}

type testLease struct {
	run   func(context.Context, Job) error
	close func() error
}

func (l *testLease) Run(ctx context.Context, job Job) error { return l.run(ctx, job) }
func (l *testLease) Close() error                           { return l.close() }

func TestServeRunsUpToCapacityAndClosesAfterRuns(t *testing.T) {
	jobs := make(chan Job, 3)
	for job := Job(1); job <= 3; job++ {
		jobs <- job
	}
	close(jobs)
	started := make(chan Job, 3)
	releaseRuns := make(chan struct{})
	closed := make(chan Job, 3)
	returned := make(chan error, 1)
	go func() {
		returned <- Serve(context.Background(), jobs, 2, func(_ context.Context, job Job) (Lease, error) {
			return &testLease{
				run:   func(context.Context, Job) error { started <- job; <-releaseRuns; return nil },
				close: func() error { closed <- job; return nil },
			}, nil
		})
	}()

	first := await(t, started, "first Run")
	second := await(t, started, "second Run")
	if first == second {
		t.Fatalf("started duplicate jobs %d", first)
	}
	select {
	case job := <-started:
		t.Fatalf("job %d started beyond capacity", job)
	default:
	}
	select {
	case job := <-closed:
		t.Fatalf("job %d closed before Run completed", job)
	default:
	}
	close(releaseRuns)
	await(t, closed, "first Close")
	await(t, closed, "second Close")
	await(t, started, "third Run after capacity release")
	// The third Run needs a release signal too; the channel is already closed.
	if err := await(t, returned, "Serve return"); err != nil {
		t.Fatalf("Serve returned %v", err)
	}
	await(t, closed, "third Close")
}

func TestServeJoinsWholeCohortBeforeClosingAndCloseHoldsCapacity(t *testing.T) {
	jobs := make(chan Job, 3)
	for job := Job(1); job <= 3; job++ {
		jobs <- job
	}
	close(jobs)
	runStarted := make(chan Job, 3)
	runRelease := make(chan struct{})
	runReturned := make(chan Job, 3)
	closeStarted := make(chan Job, 3)
	closeRelease := make(chan struct{})
	thirdStarted := make(chan struct{})
	returned := make(chan error, 1)
	go func() {
		returned <- Serve(context.Background(), jobs, 2, func(_ context.Context, job Job) (Lease, error) {
			return &testLease{
				run: func(context.Context, Job) error {
					if job == 3 {
						close(thirdStarted)
					}
					runStarted <- job
					<-runRelease
					runReturned <- job
					return nil
				},
				close: func() error {
					select {
					case <-runReturned:
					default:
						return errors.New("Close began before cohort Runs returned")
					}
					closeStarted <- job
					if job == 1 {
						<-closeRelease
					}
					return nil
				},
			}, nil
		})
	}()
	await(t, runStarted, "first Run")
	await(t, runStarted, "second Run")
	close(runRelease)
	await(t, closeStarted, "first Close")
	select {
	case <-thirdStarted:
		t.Fatal("third Run started before held Close completed")
	default:
	}
	close(closeRelease)
	await(t, closeStarted, "second Close")
	await(t, thirdStarted, "third Run")
	if err := await(t, returned, "Serve return"); err != nil {
		t.Fatalf("Serve returned %v", err)
	}
}

func TestServeInputWaitIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan Job)
	returned := make(chan error, 1)
	go func() {
		returned <- Serve(ctx, jobs, 2, func(context.Context, Job) (Lease, error) { t.Error("unexpected Open"); return nil, nil })
	}()
	cancel()
	if err := await(t, returned, "canceled Serve"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Serve error = %v, want context.Canceled", err)
	}
}

func TestServeAdmitsLaterInputWhileCapacityRemains(t *testing.T) {
	jobs := make(chan Job)
	started := make(chan Job, 2)
	release := make(chan struct{})
	returned := make(chan error, 1)
	go func() {
		returned <- Serve(context.Background(), jobs, 2, func(context.Context, Job) (Lease, error) {
			return &testLease{run: func(_ context.Context, job Job) error { started <- job; <-release; return nil }, close: func() error { return nil }}, nil
		})
	}()
	jobs <- 1
	await(t, started, "first Run")
	select {
	case <-returned:
		t.Fatal("Serve returned while input remained open and capacity was available")
	default:
	}
	jobs <- 2
	await(t, started, "second Run")
	close(release)
	close(jobs)
	if err := await(t, returned, "Serve after input close"); err != nil {
		t.Fatalf("Serve returned %v", err)
	}
}

func TestServeOpenFailureStopsAndReleasesStartedWork(t *testing.T) {
	openErr := errors.New("open failed")
	started := make(chan struct{})
	runStopped := make(chan struct{})
	closed := make(chan struct{})
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	returned := make(chan error, 1)
	var opens sync.Mutex
	count := 0
	go func() {
		returned <- Serve(context.Background(), jobs, 2, func(context.Context, Job) (Lease, error) {
			opens.Lock()
			count++
			n := count
			opens.Unlock()
			if n == 1 {
				return &testLease{
					run: func(ctx context.Context, _ Job) error {
						close(started)
						<-ctx.Done()
						close(runStopped)
						return ctx.Err()
					},
					close: func() error { close(closed); return nil },
				}, nil
			}
			<-started
			return nil, openErr
		})
	}()
	if err := await(t, returned, "Serve after Open failure"); !errors.Is(err, openErr) {
		t.Fatalf("Serve error = %v, want open failure", err)
	}
	select {
	case <-runStopped:
	default:
		t.Fatal("Run did not observe stop before Serve returned")
	}
	select {
	case <-closed:
	default:
		t.Fatal("successful lease was not closed")
	}
}

func TestServeRetainsRunAndCloseErrors(t *testing.T) {
	runErr := errors.New("run failed")
	closeErr := errors.New("close failed")
	jobs := make(chan Job, 1)
	jobs <- 1
	close(jobs)
	err := Serve(context.Background(), jobs, 1, func(context.Context, Job) (Lease, error) {
		return &testLease{run: func(context.Context, Job) error { return runErr }, close: func() error { return closeErr }}, nil
	})
	if !errors.Is(err, runErr) || !errors.Is(err, closeErr) {
		t.Fatalf("Serve error = %v, want both Run and Close failures", err)
	}
}

func TestServeRetainsIndependentOpenAndCloseErrors(t *testing.T) {
	openErr := errors.New("open failed")
	closeErr := errors.New("close failed")
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	var closes sync.Mutex
	closeCount := 0
	err := Serve(context.Background(), jobs, 2, func(_ context.Context, job Job) (Lease, error) {
		if job == 2 {
			return nil, openErr
		}
		return &testLease{run: func(context.Context, Job) error { return nil }, close: func() error { closes.Lock(); closeCount++; closes.Unlock(); return closeErr }}, nil
	})
	if !errors.Is(err, openErr) || !errors.Is(err, closeErr) {
		t.Fatalf("Serve error = %v, want Open and Close failures", err)
	}
	if closeCount != 1 {
		t.Fatalf("Close calls = %d, want 1", closeCount)
	}
}

func TestServeAlreadyCanceledOpensNothing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	opened := false
	err := Serve(ctx, make(chan Job), 2, func(context.Context, Job) (Lease, error) { opened = true; return nil, nil })
	if !errors.Is(err, context.Canceled) || opened {
		t.Fatalf("Serve error=%v opened=%v", err, opened)
	}
}
