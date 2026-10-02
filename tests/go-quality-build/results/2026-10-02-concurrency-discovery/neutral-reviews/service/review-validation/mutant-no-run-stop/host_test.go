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

func waitSignal(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func waitJob(t *testing.T, ch <-chan Job, what string) Job {
	t.Helper()
	select {
	case job := <-ch:
		return job
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
		return 0
	}
}

func TestServeCapacityAndJoinBeforeRelease(t *testing.T) {
	jobs := make(chan Job, 3)
	for i := 0; i < 3; i++ {
		jobs <- Job(i)
	}
	close(jobs)
	started := make(chan Job, 3)
	release := make(chan struct{})
	var mu sync.Mutex
	active, maxActive, closed := 0, 0, 0
	done := make(chan error, 1)
	go func() {
		done <- Serve(context.Background(), jobs, 2, func(_ context.Context, j Job) (Lease, error) {
			return &testLease{run: func(context.Context, Job) error {
				mu.Lock()
				active++
				if active > maxActive {
					maxActive = active
				}
				mu.Unlock()
				started <- j
				<-release
				mu.Lock()
				active--
				mu.Unlock()
				return nil
			}, close: func() error {
				mu.Lock()
				defer mu.Unlock()
				if active != 0 {
					return errors.New("closed before cohort joined")
				}
				closed++
				return nil
			}}, nil
		})
	}()
	waitJob(t, started, "first Run")
	waitJob(t, started, "second Run")
	select {
	case j := <-started:
		t.Fatalf("admitted job %d before capacity was released", j)
	case <-time.After(40 * time.Millisecond):
	}
	close(release)
	waitJob(t, started, "third Run")
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve did not return")
	}
	mu.Lock()
	defer mu.Unlock()
	if maxActive != 2 || closed != 3 {
		t.Fatalf("max active=%d closed=%d, want 2 and 3", maxActive, closed)
	}
}

func TestServeJoinsIndependentFailures(t *testing.T) {
	runErr, closeErr := errors.New("run"), errors.New("close")
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	var closes int
	err := Serve(context.Background(), jobs, 2, func(_ context.Context, j Job) (Lease, error) {
		return &testLease{run: func(context.Context, Job) error {
			if j == 1 {
				return runErr
			}
			return nil
		}, close: func() error { closes++; return closeErr }}, nil
	})
	for _, want := range []error{runErr, closeErr} {
		if !errors.Is(err, want) {
			t.Errorf("Serve error %v does not include %v", err, want)
		}
	}
	if closes != 2 {
		t.Fatalf("Close calls=%d, want 2", closes)
	}
}

func TestServeRetainsOpenFailure(t *testing.T) {
	openErr := errors.New("open")
	jobs := make(chan Job, 1)
	jobs <- 1
	close(jobs)
	err := Serve(context.Background(), jobs, 2, func(context.Context, Job) (Lease, error) { return nil, openErr })
	if !errors.Is(err, openErr) {
		t.Fatalf("Serve error=%v, want Open error", err)
	}
}

func TestServeCancellationWhileWaitingForJobs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan Job)
	called := false
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, jobs, 2, func(context.Context, Job) (Lease, error) { called = true; return nil, nil })
	}()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Serve error=%v, want cancellation", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Serve did not stop waiting for input")
	}
	if called {
		t.Fatal("Open called after cancellation")
	}
}

func TestServeAlreadyCanceledDoesNotOpen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	jobs := make(chan Job, 1)
	jobs <- 1
	err := Serve(ctx, jobs, 1, func(context.Context, Job) (Lease, error) { t.Fatal("Open called"); return nil, nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Serve error=%v, want cancellation", err)
	}
}
