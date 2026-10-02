package sweeper_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/sweeper"
)

// This assignment checks the supported function type at a consumer boundary.
var _ func(context.Context, time.Duration, func(context.Context) error, func() error) error = sweeper.Run

const waitLimit = 2 * time.Second

type runCall struct {
	cancel context.CancelFunc
	done   chan struct{}
	err    error // Read only after done is closed.
}

func startRun(t *testing.T, parent context.Context, interval time.Duration, sweep func(context.Context) error, release func() error) *runCall {
	t.Helper()
	ctx, cancel := context.WithCancel(parent)
	r := &runCall{cancel: cancel, done: make(chan struct{})}
	go func() {
		r.err = sweeper.Run(ctx, interval, sweep, release)
		close(r.done)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-r.done:
		case <-time.After(waitLimit):
			t.Error("Run did not finish during test cleanup")
		}
	})
	return r
}

func (r *runCall) result(t *testing.T) error {
	t.Helper()
	select {
	case <-r.done:
		return r.err
	case <-time.After(waitLimit):
		t.Fatal("Run did not finish")
		return nil
	}
}

func receive[T any](t *testing.T, event <-chan T, r *runCall, label string) T {
	t.Helper()
	select {
	case value := <-event:
		return value
	case <-r.done:
		t.Fatalf("Run ended before %s: %v", label, r.err)
	case <-time.After(waitLimit):
		t.Fatalf("timed out waiting for %s", label)
	}
	var zero T
	return zero
}

func unblocker(gate chan struct{}) func() {
	var once sync.Once
	return func() { once.Do(func() { close(gate) }) }
}

func TestRunRejectsNonpositiveIntervals(t *testing.T) {
	for _, interval := range []time.Duration{0, -time.Nanosecond, -time.Hour} {
		t.Run(interval.String(), func(t *testing.T) {
			var sweeps, releases atomic.Int32
			r := startRun(t, context.Background(), interval, func(context.Context) error {
				sweeps.Add(1)
				return nil
			}, func() error {
				releases.Add(1)
				return nil
			})
			if err := r.result(t); err == nil {
				t.Error("Run accepted a nonpositive interval")
			}
			if sweeps.Load() != 0 || releases.Load() != 0 {
				t.Errorf("sweeps = %d, releases = %d; want neither", sweeps.Load(), releases.Load())
			}
		})
	}
}

func TestRunAlreadyCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var sweeps, releases atomic.Int32
	r := startRun(t, ctx, time.Hour, func(context.Context) error {
		sweeps.Add(1)
		return nil
	}, func() error {
		releases.Add(1)
		return nil
	})
	if err := r.result(t); err != nil {
		t.Fatalf("Run = %v, want nil for parent cancellation", err)
	}
	if sweeps.Load() != 0 || releases.Load() != 1 {
		t.Errorf("sweeps = %d, releases = %d; want 0, 1", sweeps.Load(), releases.Load())
	}
}

func TestRunStartsImmediatelyAndCancelsIntervalWait(t *testing.T) {
	started := make(chan struct{}, 1)
	var sweeps, releases atomic.Int32
	r := startRun(t, context.Background(), time.Hour, func(context.Context) error {
		sweeps.Add(1)
		started <- struct{}{}
		return nil
	}, func() error {
		releases.Add(1)
		return nil
	})
	receive(t, started, r, "immediate sweep")
	// A completed callback must not terminate the host. The hour-long interval
	// also makes a delayed first call or an uncancelable wait fail promptly.
	select {
	case <-r.done:
		t.Fatalf("Run ended after its first successful sweep: %v", r.err)
	case <-time.After(20 * time.Millisecond):
	}
	r.cancel()
	if err := r.result(t); err != nil {
		t.Fatalf("Run = %v, want nil", err)
	}
	if sweeps.Load() != 1 || releases.Load() != 1 {
		t.Errorf("sweeps = %d, releases = %d; want 1, 1", sweeps.Load(), releases.Load())
	}
}

func TestRunWaitsFullIntervalAfterEachCompletion(t *testing.T) {
	const interval = 40 * time.Millisecond
	type cycle struct {
		number int32
		start  time.Time
	}
	cycles := make(chan cycle, 16)
	completed := make(chan time.Time, 2)
	firstGate := make(chan struct{})
	unblock := unblocker(firstGate)
	defer unblock()
	var calls, active, releases atomic.Int32
	var overlap atomic.Bool
	r := startRun(t, context.Background(), interval, func(ctx context.Context) error {
		if active.Add(1) != 1 {
			overlap.Store(true)
		}
		defer active.Add(-1)
		number := calls.Add(1)
		select {
		case cycles <- cycle{number: number, start: time.Now()}:
		case <-ctx.Done():
			return nil
		}
		switch number {
		case 1:
			select {
			case <-firstGate:
			case <-ctx.Done():
				return nil
			}
			completed <- time.Now()
		case 2:
			completed <- time.Now()
		default:
			<-ctx.Done()
		}
		return nil
	}, func() error { releases.Add(1); return nil })
	if first := receive(t, cycles, r, "first cycle"); first.number != 1 {
		t.Fatalf("first cycle = %d, want 1", first.number)
	}
	// Hold the callback beyond the interval. A ticker or overlapping scheduler
	// must not turn elapsed callback time into an immediately pending cycle.
	select {
	case cycle := <-cycles:
		t.Fatalf("cycle %d overlapped the blocked first callback", cycle.number)
	case <-r.done:
		t.Fatalf("Run ended while the first callback was blocked: %v", r.err)
	case <-time.After(2 * interval):
	}
	unblock()
	for _, wantNumber := range []int32{2, 3} {
		finishedAt := receive(t, completed, r, "previous callback completion")
		next := receive(t, cycles, r, "next cycle")
		if next.number != wantNumber {
			t.Fatalf("cycle = %d, want %d", next.number, wantNumber)
		}
		if elapsed := next.start.Sub(finishedAt); elapsed < interval {
			t.Errorf("cycle %d started %v after completion, want at least %v", next.number, elapsed, interval)
		}
	}
	r.cancel()
	if err := r.result(t); err != nil {
		t.Fatalf("Run = %v, want nil", err)
	}
	if overlap.Load() || active.Load() != 0 || calls.Load() != 3 || releases.Load() != 1 {
		t.Errorf("overlap = %v, active = %d, calls = %d, releases = %d; want false, 0, 3, 1", overlap.Load(), active.Load(), calls.Load(), releases.Load())
	}
}

func TestRunCancellationWaitsForCallbackCleanup(t *testing.T) {
	independentErr := errors.New("cleanup inside callback failed")
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "successful callback cleanup"},
		{name: "independent callback failure", err: independentErr},
		{name: "joined cancellation and callback failure", err: errors.Join(context.Canceled, independentErr)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			started := make(chan struct{})
			canceled := make(chan struct{})
			cleanupGate := make(chan struct{})
			cleaned := make(chan struct{})
			released := make(chan struct{})
			unblock := unblocker(cleanupGate)
			defer unblock()
			var releases atomic.Int32
			var earlyRelease atomic.Bool
			r := startRun(t, context.Background(), time.Hour, func(ctx context.Context) error {
				close(started)
				<-ctx.Done()
				close(canceled)
				<-cleanupGate
				close(cleaned)
				return tc.err
			}, func() error {
				select {
				case <-cleaned:
				default:
					earlyRelease.Store(true)
				}
				releases.Add(1)
				close(released)
				return nil
			})
			receive(t, started, r, "callback start")
			r.cancel()
			receive(t, canceled, r, "callback context cancellation")
			select {
			case <-released:
				t.Fatal("release ran before callback cleanup finished")
			case <-r.done:
				t.Fatal("Run returned before callback cleanup finished")
			case <-time.After(20 * time.Millisecond):
			}
			unblock()
			err := r.result(t)
			if tc.err == nil && err != nil {
				t.Errorf("Run = %v, want nil", err)
			}
			if tc.err != nil && !errors.Is(err, independentErr) {
				t.Errorf("Run = %v, want independent callback cause", err)
			}
			if errors.Is(tc.err, context.Canceled) && !errors.Is(err, context.Canceled) {
				t.Errorf("Run = %v, lost cancellation cause returned by callback", err)
			}
			if earlyRelease.Load() || releases.Load() != 1 {
				t.Errorf("early release = %v, releases = %d; want false, 1", earlyRelease.Load(), releases.Load())
			}
		})
	}
}

func TestRunErrorCombinations(t *testing.T) {
	workErr := errors.New("sweep failed")
	releaseErr := errors.New("release failed")
	for _, tc := range []struct {
		name          string
		work, release error
	}{
		{name: "success"},
		{name: "work fails", work: workErr},
		{name: "release fails", release: releaseErr},
		{name: "both fail", work: workErr, release: releaseErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parent, cancelParent := context.WithCancel(context.Background())
			defer cancelParent()
			var releases atomic.Int32
			r := startRun(t, parent, time.Hour, func(context.Context) error {
				if tc.work == nil {
					cancelParent()
				}
				return tc.work
			}, func() error { releases.Add(1); return tc.release })
			err := r.result(t)
			if tc.work == nil && tc.release == nil && err != nil {
				t.Errorf("Run = %v, want nil", err)
			}
			for _, want := range []error{tc.work, tc.release} {
				if want != nil && !errors.Is(err, want) {
					t.Errorf("Run = %v, lost cause %v", err, want)
				}
			}
			if errors.Is(err, workErr) != (tc.work != nil) || errors.Is(err, releaseErr) != (tc.release != nil) {
				t.Errorf("Run = %v, included an unexpected work or release cause", err)
			}
			if errors.Is(err, context.Canceled) {
				t.Errorf("Run = %v, reported parent cancellation as a worker failure", err)
			}
			if releases.Load() != 1 {
				t.Errorf("releases = %d, want 1", releases.Load())
			}
		})
	}
}

func TestSweepFailureReleases(t *testing.T) {
	want := errors.New("sweep failed")
	var releases atomic.Int32
	var callbackCtx context.Context
	var canceledAtRelease atomic.Bool
	r := startRun(t, context.Background(), time.Hour, func(ctx context.Context) error {
		callbackCtx = ctx
		return want
	}, func() error {
		canceledAtRelease.Store(errors.Is(callbackCtx.Err(), context.Canceled))
		releases.Add(1)
		return nil
	})
	if err := r.result(t); !errors.Is(err, want) {
		t.Fatalf("Run = %v, want sweep error identity", err)
	}
	if releases.Load() != 1 || !canceledAtRelease.Load() {
		t.Errorf("releases = %d, context canceled at release = %v; want 1, true", releases.Load(), canceledAtRelease.Load())
	}
}

func TestRunWaitsForReleaseToFinish(t *testing.T) {
	workErr := errors.New("sweep failed")
	releaseErr := errors.New("release failed")
	releaseStarted := make(chan struct{})
	releaseGate := make(chan struct{})
	unblock := unblocker(releaseGate)
	defer unblock()
	r := startRun(t, context.Background(), time.Hour, func(context.Context) error {
		return workErr
	}, func() error {
		close(releaseStarted)
		<-releaseGate
		return releaseErr
	})
	receive(t, releaseStarted, r, "release start")
	select {
	case <-r.done:
		t.Fatalf("Run returned before release finished: %v", r.err)
	case <-time.After(20 * time.Millisecond):
	}
	unblock()
	err := r.result(t)
	if !errors.Is(err, workErr) || !errors.Is(err, releaseErr) {
		t.Errorf("Run = %v, want both causes after release completion", err)
	}
}

type detailedFailure struct {
	details []string
	cause   error
}

func (e detailedFailure) Error() string { return "sweep failed" }
func (e detailedFailure) Unwrap() error { return e.cause }

func TestRunStopsAfterLaterFailure(t *testing.T) {
	want := errors.New("third sweep failed")
	var calls, releases atomic.Int32
	r := startRun(t, context.Background(), time.Millisecond, func(context.Context) error {
		if calls.Add(1) == 3 {
			return detailedFailure{details: []string{"third", "cycle"}, cause: want}
		}
		return nil
	}, func() error { releases.Add(1); return nil })
	err := r.result(t)
	var got detailedFailure
	if !errors.Is(err, want) || !errors.As(err, &got) {
		t.Fatalf("Run = %v, want preserved error cause and type", err)
	}
	if len(got.details) != 2 || got.details[0] != "third" || got.details[1] != "cycle" {
		t.Errorf("error details = %v, want [third cycle]", got.details)
	}
	if calls.Load() != 3 || releases.Load() != 1 {
		t.Errorf("calls = %d, releases = %d; want 3, 1", calls.Load(), releases.Load())
	}
}

func TestRunInstancesStopIndependently(t *testing.T) {
	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})
	secondRecurred := make(chan struct{})
	secondGate := make(chan struct{})
	unblock := unblocker(secondGate)
	defer unblock()
	var firstReleases, secondReleases, secondCalls atomic.Int32
	first := startRun(t, context.Background(), time.Hour, func(ctx context.Context) error {
		close(firstStarted)
		<-ctx.Done()
		return nil
	}, func() error { firstReleases.Add(1); return nil })
	second := startRun(t, context.Background(), time.Millisecond, func(ctx context.Context) error {
		if secondCalls.Add(1) == 1 {
			close(secondStarted)
			select {
			case <-secondGate:
			case <-ctx.Done():
			}
			return nil
		}
		close(secondRecurred)
		<-ctx.Done()
		return nil
	}, func() error { secondReleases.Add(1); return nil })
	receive(t, firstStarted, first, "first instance start")
	receive(t, secondStarted, second, "second instance start")
	first.cancel()
	if err := first.result(t); err != nil {
		t.Fatalf("first Run = %v, want nil", err)
	}
	if firstReleases.Load() != 1 || secondReleases.Load() != 0 {
		t.Errorf("first releases = %d, second releases = %d; want 1, 0", firstReleases.Load(), secondReleases.Load())
	}
	unblock()
	receive(t, secondRecurred, second, "second recurrence after first joined")
	second.cancel()
	if err := second.result(t); err != nil {
		t.Fatalf("second Run = %v, want nil", err)
	}
	if secondCalls.Load() != 2 || secondReleases.Load() != 1 {
		t.Errorf("second calls = %d, releases = %d; want 2, 1", secondCalls.Load(), secondReleases.Load())
	}
}
