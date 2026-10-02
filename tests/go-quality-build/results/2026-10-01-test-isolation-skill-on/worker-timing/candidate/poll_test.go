package poller_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/poller"
)

func TestCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	var calls atomic.Int32
	run := startPoll(t, ctx, time.Hour, func(context.Context) error {
		if calls.Add(1) == 1 {
			close(started)
		}
		return nil
	}, cancel)

	// A one-hour interval also checks that the first callback starts immediately.
	receive(t, started, "first callback")
	cancel()
	if err := run.wait(t); err != nil {
		t.Fatalf("Poll after cancellation = %v, want nil", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("callback count = %d, want 1", got)
	}
}

func TestCallbackError(t *testing.T) {
	t.Parallel()
	want := errors.New("poll failed")
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	run := startPoll(t, ctx, time.Hour, func(context.Context) error {
		calls.Add(1)
		return want
	}, cancel)
	if err := run.wait(t); err != want {
		t.Fatalf("Poll = %v, want original callback error %v", err, want)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("callback count = %d, want 1", got)
	}
}

func TestAlreadyCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var calls atomic.Int32
	run := startPoll(t, ctx, time.Hour, func(context.Context) error {
		calls.Add(1)
		return nil
	}, cancel)
	if err := run.wait(t); err != nil {
		t.Fatalf("Poll with canceled context = %v, want nil", err)
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("callback count = %d, want 0", got)
	}
}

func TestInvalidInterval(t *testing.T) {
	t.Parallel()
	for _, interval := range []time.Duration{0, -time.Nanosecond} {
		t.Run(interval.String(), func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			var calls atomic.Int32
			run := startPoll(t, ctx, interval, func(context.Context) error {
				calls.Add(1)
				return nil
			}, cancel)
			if err := run.wait(t); err == nil {
				t.Fatal("Poll with invalid interval returned nil")
			}
			if got := calls.Load(); got != 0 {
				t.Fatalf("callback count = %d, want 0", got)
			}
		})
	}
}

func TestRecurrence(t *testing.T) {
	t.Parallel()
	const interval = 40 * time.Millisecond
	type start struct {
		call   int32
		active int32
		at     time.Time
	}
	ctx, cancel := context.WithCancel(context.Background())
	starts := make(chan start, 3)
	finishes := make(chan time.Time, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	stop := func() {
		cancel()
		releaseOnce.Do(func() { close(release) })
	}
	var calls, active atomic.Int32
	want := errors.New("third callback finished")
	run := startPoll(t, ctx, interval, func(ctx context.Context) error {
		call := calls.Add(1)
		concurrent := active.Add(1)
		defer func() {
			active.Add(-1)
			if call < 3 {
				// Sample at the end of callback work and cleanup. Releasing
				// the gate only grants permission to finish.
				select {
				case finishes <- time.Now():
				case <-ctx.Done():
				}
			}
		}()
		select {
		case starts <- start{call: call, active: concurrent, at: time.Now()}:
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
		if call >= 3 {
			return want
		}
		return nil
	}, stop)

	first := receive(t, starts, "first callback start")
	if first.call != 1 || first.active != 1 {
		t.Fatalf("first start = %+v, want call 1 with one active callback", first)
	}
	// Startup is acknowledged before this observation. Holding the first
	// callback longer than an interval exposes overlapping/start-relative cycles.
	hold := time.NewTimer(2 * interval)
	defer hold.Stop()
	select {
	case next := <-starts:
		t.Fatalf("callback %d started while callback 1 was blocked", next.call)
	case <-hold.C:
	}
	releaseOnce.Do(func() { close(release) })
	for call := int32(2); call <= 3; call++ {
		finished := receive(t, finishes, "preceding callback completion")
		next := receive(t, starts, "recurring callback start")
		if next.call != call || next.active != 1 {
			t.Fatalf("start = %+v, want call %d with one active callback", next, call)
		}
		// Monotonic elapsed time checks the lower bound only. Late scheduling
		// is allowed; the event timeout diagnoses stalls rather than machine speed.
		if elapsed := next.at.Sub(finished); elapsed < interval {
			t.Fatalf("callback %d started %v after completion, want at least %v", call, elapsed, interval)
		}
	}
	if err := run.wait(t); err != want {
		t.Fatalf("Poll = %v, want %v", err, want)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("callback count = %d, want 3", got)
	}
}

func TestCancellationDuringCallbackCleanup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan context.Context, 1)
	cleanupStarted := make(chan struct{})
	cleanupFinished := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	var calls atomic.Int32
	stop := func() {
		cancel()
		releaseOnce.Do(func() { close(release) })
	}
	run := startPoll(t, ctx, time.Hour, func(callbackCtx context.Context) error {
		if calls.Add(1) != 1 {
			return errors.New("unexpected subsequent callback")
		}
		defer func() {
			close(cleanupStarted)
			<-release
			close(cleanupFinished)
		}()
		started <- callbackCtx
		<-callbackCtx.Done()
		return nil
	}, stop)
	if got := receive(t, started, "active callback"); got != ctx {
		t.Fatal("callback did not receive the caller's context")
	}
	cancel()
	receive(t, cleanupStarted, "callback cleanup start")
	// The callback has acknowledged cancellation and blocked in cleanup.
	observation := time.NewTimer(20 * time.Millisecond)
	defer observation.Stop()
	select {
	case <-run.done:
		t.Fatal("Poll returned while callback cleanup was blocked")
	case <-observation.C:
	}
	releaseOnce.Do(func() { close(release) })
	if err := run.wait(t); err != nil {
		t.Fatalf("Poll after active cancellation = %v, want nil", err)
	}
	select {
	case <-cleanupFinished:
	default:
		t.Fatal("Poll returned before callback cleanup finished")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("callback count = %d, want 1", got)
	}
}

func TestCallbackErrorConcurrentWithCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	want := errors.New("callback failed during cancellation")
	var calls atomic.Int32
	run := startPoll(t, ctx, time.Hour, func(callbackCtx context.Context) error {
		if calls.Add(1) == 1 {
			close(started)
		}
		<-callbackCtx.Done()
		return want
	}, cancel)
	receive(t, started, "callback before cancellation")
	cancel()
	if err := run.wait(t); err != want {
		t.Fatalf("Poll = %v, want original callback error %v", err, want)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("callback count = %d, want 1", got)
	}
}

func TestFixtureTeardown(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "callback.txt")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	var run *pollRun
	t.Cleanup(func() {
		if run != nil {
			select {
			case <-run.done:
			default:
				t.Error("callback still owns fixture; cannot safely close it")
				return
			}
		}
		if err := file.Close(); err != nil {
			t.Errorf("close caller-owned fixture: %v", err)
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	run = startPoll(t, ctx, time.Hour, func(callbackCtx context.Context) (err error) {
		defer func() {
			_, err = file.WriteString("callback cleanup finished\n")
		}()
		close(started)
		<-callbackCtx.Done()
		return nil
	}, cancel)
	receive(t, started, "fixture borrower")
	cancel()
	if err := run.wait(t); err != nil {
		t.Fatalf("Poll fixture callback = %v, want nil", err)
	}
	if _, err := file.Stat(); err != nil {
		t.Fatalf("caller-owned fixture unavailable after Poll: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "callback cleanup finished\n" {
		t.Fatalf("fixture contents = %q, want callback cleanup marker", got)
	}
}

func TestConcurrentInvocations(t *testing.T) {
	t.Parallel()
	firstCtx, cancelFirst := context.WithCancel(context.Background())
	secondCtx, cancelSecond := context.WithCancel(context.Background())
	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})
	var firstCalls, secondCalls atomic.Int32
	first := startPoll(t, firstCtx, time.Hour, func(ctx context.Context) error {
		if firstCalls.Add(1) == 1 {
			close(firstStarted)
		}
		<-ctx.Done()
		return nil
	}, cancelFirst)
	second := startPoll(t, secondCtx, time.Hour, func(ctx context.Context) error {
		if secondCalls.Add(1) == 1 {
			close(secondStarted)
		}
		<-ctx.Done()
		return nil
	}, cancelSecond)
	receive(t, firstStarted, "first invocation startup")
	receive(t, secondStarted, "second invocation startup")
	cancelFirst()
	if err := first.wait(t); err != nil {
		t.Fatalf("first Poll = %v, want nil", err)
	}
	// Both invocations acknowledged startup before cancellation. The second
	// remains active until its own caller cancels it.
	observation := time.NewTimer(20 * time.Millisecond)
	defer observation.Stop()
	select {
	case <-second.done:
		t.Fatal("canceling the first invocation stopped the second")
	case <-observation.C:
	}
	if err := secondCtx.Err(); err != nil {
		t.Fatalf("second caller context = %v, want active", err)
	}
	cancelSecond()
	if err := second.wait(t); err != nil {
		t.Fatalf("second Poll = %v, want nil", err)
	}
	if firstCalls.Load() != 1 || secondCalls.Load() != 1 {
		t.Fatalf("callback counts = (%d, %d), want (1, 1)", firstCalls.Load(), secondCalls.Load())
	}
}

func TestIndependentInvocations(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"alpha", "beta"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Each selectable child owns its context, state and path. No parent
			// assertion assumes that its sibling was selected or ran.
			path := filepath.Join(t.TempDir(), "invocation.txt")
			ctx, cancel := context.WithCancel(context.Background())
			started := make(chan struct{})
			var calls atomic.Int32
			want := errors.New(name + " completed")
			run := startPoll(t, ctx, time.Hour, func(callbackCtx context.Context) error {
				if calls.Add(1) != 1 {
					return errors.New("unexpected subsequent callback")
				}
				close(started)
				<-callbackCtx.Done()
				if err := os.WriteFile(path, []byte(name), 0o600); err != nil {
					return err
				}
				return want
			}, cancel)
			receive(t, started, name+" callback")
			cancel()
			if err := run.wait(t); err != want {
				t.Fatalf("Poll = %v, want invocation error %v", err, want)
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != name {
				t.Fatalf("fixture contents = %q, want %q", got, name)
			}
			if got := calls.Load(); got != 1 {
				t.Fatalf("callback count = %d, want 1", got)
			}
		})
	}
}

type pollRun struct {
	done chan struct{}
	err  error // Published by closing done; read only after joining.
}

func startPoll(t *testing.T, ctx context.Context, interval time.Duration, callback func(context.Context) error, stop func()) *pollRun {
	t.Helper()
	run := &pollRun{done: make(chan struct{})}
	// Register ownership before launch. stop also releases callback gates so
	// an early Fatal cannot leave a worker waiting on the test goroutine.
	t.Cleanup(func() {
		stop()
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		select {
		case <-run.done:
		case <-timer.C:
			t.Error("Poll did not join during cleanup")
		}
	})
	go func() {
		run.err = poller.Poll(ctx, interval, callback)
		close(run.done)
	}()
	return run
}

func (run *pollRun) wait(t *testing.T) error {
	t.Helper()
	receive(t, run.done, "Poll completion")
	return run.err
}

func receive[T any](t *testing.T, ch <-chan T, operation string) T {
	t.Helper()
	timer := time.NewTimer(5 * time.Second)
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
