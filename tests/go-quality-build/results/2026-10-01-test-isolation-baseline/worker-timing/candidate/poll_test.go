package poller_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/poller"
)

const waitLimit = 2 * time.Second

// A closed done channel lets both the test and its cleanup join the same run.
type pollRun struct {
	done chan struct{}
	err  error
}

func startPoll(t *testing.T, ctx context.Context, cancel context.CancelFunc, interval time.Duration, callback func(context.Context) error) *pollRun {
	t.Helper()
	run := &pollRun{done: make(chan struct{})}
	go func() {
		run.err = poller.Poll(ctx, interval, callback)
		close(run.done)
	}()
	t.Cleanup(func() {
		cancel()
		await(t, run.done, "Poll teardown")
	})
	return run
}

func await[T any](t *testing.T, ch <-chan T, operation string) T {
	t.Helper()
	timer := time.NewTimer(waitLimit)
	defer timer.Stop()
	select {
	case value := <-ch:
		return value
	case <-timer.C:
		t.Fatalf("timed out waiting for %s", operation)
		var zero T
		return zero
	}
}

func (run *pollRun) wait(t *testing.T) error {
	t.Helper()
	await(t, run.done, "Poll return")
	return run.err
}

func TestCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	returned := make(chan struct{})
	var calls atomic.Int32
	run := startPoll(t, ctx, cancel, time.Hour, func(context.Context) error {
		if calls.Add(1) == 1 {
			close(returned)
		}
		return nil
	})

	// The first callback must begin promptly, even with an hour-long interval.
	await(t, returned, "first callback")
	cancel()
	if err := run.wait(t); err != nil {
		t.Fatalf("Poll after cancellation = %v, want nil", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("callback calls = %d, want 1", got)
	}
}

func TestAlreadyCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var calls atomic.Int32
	run := startPoll(t, ctx, cancel, time.Hour, func(context.Context) error {
		calls.Add(1)
		return errors.New("unexpected callback")
	})
	if err := run.wait(t); err != nil {
		t.Fatalf("Poll with canceled context = %v, want nil", err)
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("callback calls = %d, want 0", got)
	}
}

func TestInvalidInterval(t *testing.T) {
	t.Parallel()
	for _, interval := range []time.Duration{0, -time.Nanosecond, -time.Second} {
		t.Run(interval.String(), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			var calls atomic.Int32
			run := startPoll(t, ctx, cancel, interval, func(context.Context) error {
				calls.Add(1)
				return errors.New("unexpected callback")
			})
			if err := run.wait(t); err == nil {
				t.Fatal("Poll with nonpositive interval returned nil")
			}
			if got := calls.Load(); got != 0 {
				t.Fatalf("callback calls = %d, want 0", got)
			}
		})
	}
}

func TestCallbackError(t *testing.T) {
	t.Parallel()
	want := errors.New("poll failed")
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	run := startPoll(t, ctx, cancel, time.Hour, func(context.Context) error {
		calls.Add(1)
		return want
	})
	if err := run.wait(t); err != want || !errors.Is(err, want) {
		t.Fatalf("Poll = %v, want original callback error %v", err, want)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("callback calls = %d, want 1", got)
	}
}

func TestCallbackErrorDuringCancellation(t *testing.T) {
	t.Parallel()
	want := errors.New("callback failed during cancellation")
	ctx, cancel := context.WithCancel(context.Background())
	entered := make(chan struct{})
	run := startPoll(t, ctx, cancel, time.Hour, func(ctx context.Context) error {
		close(entered)
		<-ctx.Done()
		return want
	})
	await(t, entered, "callback entry")
	cancel()
	if err := run.wait(t); err != want {
		t.Fatalf("Poll = %v, want original callback error %v", err, want)
	}
}

func TestCallerContext(t *testing.T) {
	t.Parallel()
	type contextKey struct{}
	parent := context.WithValue(context.Background(), contextKey{}, "caller value")
	ctx, cancel := context.WithCancel(parent)
	received := make(chan context.Context, 1)
	want := errors.New("finished")
	run := startPoll(t, ctx, cancel, time.Hour, func(callbackCtx context.Context) error {
		received <- callbackCtx
		return want
	})
	if got := await(t, received, "callback context"); got != ctx {
		t.Errorf("callback context = %v, want caller context %v", got, ctx)
	}
	if err := run.wait(t); err != want {
		t.Fatalf("Poll = %v, want %v", err, want)
	}
}

func TestRecurrenceAfterCompletion(t *testing.T) {
	t.Parallel()
	const interval = 20 * time.Millisecond
	type event struct {
		call   int32
		start  bool
		at     time.Time
		active int32
	}
	events := make(chan event, 8)
	release := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	var calls, active atomic.Int32
	want := errors.New("three callbacks complete")
	run := startPoll(t, ctx, cancel, interval, func(ctx context.Context) error {
		call := calls.Add(1)
		current := active.Add(1)
		defer active.Add(-1)
		select {
		case events <- event{call: call, start: true, at: time.Now(), active: current}:
		case <-ctx.Done():
			return nil
		}
		if call == 1 {
			select {
			case <-release:
			case <-ctx.Done():
				return nil
			}
		}
		if call == 3 {
			return want
		}
		select {
		case events <- event{call: call, at: time.Now()}:
		case <-ctx.Done():
		}
		return nil
	})

	first := await(t, events, "first callback entry")
	if first.call != 1 || !first.start || first.active != 1 {
		t.Fatalf("first callback event = %+v", first)
	}
	// Hold the callback longer than the interval to expose tick-based or
	// overlapping recurrence. This timer measures a contract, not readiness.
	hold := time.NewTimer(2 * interval)
	defer hold.Stop()
	select {
	case got := <-events:
		t.Fatalf("callback event while first callback is blocked: %+v", got)
	case <-hold.C:
	}
	close(release)
	for call := int32(2); call <= 3; call++ {
		finished := await(t, events, "preceding callback completion")
		started := await(t, events, "recurring callback entry")
		if finished.start || finished.call != call-1 {
			t.Fatalf("completion event = %+v, want callback %d completion", finished, call-1)
		}
		if !started.start || started.call != call || started.active != 1 {
			t.Fatalf("start event = %+v, want sequential callback %d", started, call)
		}
		if elapsed := started.at.Sub(finished.at); elapsed < interval {
			t.Errorf("callback %d started %v after preceding completion, want at least %v", call, elapsed, interval)
		}
	}
	if err := run.wait(t); err != want {
		t.Fatalf("Poll = %v, want %v", err, want)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("callback calls = %d, want 3", got)
	}
}

func TestCancellationWaitsForCleanupAndPreservesCallerResource(t *testing.T) {
	t.Parallel()
	file, err := os.CreateTemp(t.TempDir(), "poll-resource")
	if err != nil {
		t.Fatal(err)
	}
	// Registered before Poll's cleanup so the run is joined before closing
	// the caller-owned fixture, including when an assertion fails.
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("close caller resource: %v", err)
		}
	})
	if _, err := file.WriteString("x"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	entered := make(chan error, 1)
	canceled := make(chan struct{})
	cleaned := make(chan struct{})
	releaseCleanup := make(chan struct{})
	var releaseOnce sync.Once
	run := startPoll(t, ctx, cancel, time.Hour, func(ctx context.Context) error {
		defer func() {
			<-releaseCleanup
			close(cleaned)
		}()
		var data [1]byte
		_, err := file.ReadAt(data[:], 0)
		entered <- err
		if err != nil {
			return err
		}
		<-ctx.Done()
		close(canceled)
		return nil
	})
	t.Cleanup(func() { releaseOnce.Do(func() { close(releaseCleanup) }) })
	if err := await(t, entered, "callback resource use"); err != nil {
		t.Fatalf("callback read caller resource: %v", err)
	}
	cancel()
	await(t, canceled, "callback cancellation")
	// Callback entry and cancellation are both observed before this bounded
	// negative check. The deferred cleanup remains deliberately blocked.
	observation := time.NewTimer(25 * time.Millisecond)
	defer observation.Stop()
	select {
	case <-run.done:
		t.Fatal("Poll returned before callback cleanup completed")
	case <-observation.C:
	}
	releaseOnce.Do(func() { close(releaseCleanup) })
	if err := run.wait(t); err != nil {
		t.Fatalf("Poll after cancellation = %v, want nil", err)
	}
	select {
	case <-cleaned:
	default:
		t.Fatal("Poll returned before deferred cleanup signaled completion")
	}
	if _, err := file.Stat(); err != nil {
		t.Fatalf("caller resource unusable after Poll returned: %v", err)
	}
}

func TestIndependentInvocations(t *testing.T) {
	t.Parallel()
	ctxA, cancelA := context.WithCancel(context.Background())
	ctxB, cancelB := context.WithCancel(context.Background())
	enteredA := make(chan struct{})
	canceledA := make(chan struct{})
	releaseA := make(chan struct{})
	var releaseAOnce sync.Once
	runA := startPoll(t, ctxA, cancelA, time.Hour, func(ctx context.Context) error {
		defer func() { <-releaseA }()
		close(enteredA)
		<-ctx.Done()
		close(canceledA)
		return nil
	})
	t.Cleanup(func() { releaseAOnce.Do(func() { close(releaseA) }) })

	enteredB := make(chan int32, 2)
	releaseB := make(chan struct{})
	var callsB atomic.Int32
	want := errors.New("second invocation complete")
	runB := startPoll(t, ctxB, cancelB, 5*time.Millisecond, func(ctx context.Context) error {
		call := callsB.Add(1)
		select {
		case enteredB <- call:
		case <-ctx.Done():
			return nil
		}
		if call == 1 {
			select {
			case <-releaseB:
			case <-ctx.Done():
			}
			return nil
		}
		return want
	})
	await(t, enteredA, "first invocation entry")
	if got := await(t, enteredB, "second invocation entry"); got != 1 {
		t.Fatalf("second invocation callback = %d, want 1", got)
	}
	cancelA()
	await(t, canceledA, "first invocation cancellation")
	close(releaseB)
	if got := await(t, enteredB, "second invocation recurrence"); got != 2 {
		t.Fatalf("second invocation callback = %d, want 2", got)
	}
	if err := runB.wait(t); err != want {
		t.Fatalf("second Poll = %v, want %v", err, want)
	}
	if ctxB.Err() != nil {
		t.Fatalf("second invocation context canceled: %v", ctxB.Err())
	}
	select {
	case <-runA.done:
		t.Fatal("first Poll returned while its callback cleanup remained blocked")
	default:
	}
	releaseAOnce.Do(func() { close(releaseA) })
	if err := runA.wait(t); err != nil {
		t.Fatalf("first Poll after cancellation = %v, want nil", err)
	}
}
