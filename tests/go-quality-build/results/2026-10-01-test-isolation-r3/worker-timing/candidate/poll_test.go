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

const testDeadline = 3 * time.Second

type pollRun struct {
	result   <-chan error
	finished <-chan struct{}
}

func startPoll(t *testing.T, ctx context.Context, cancel context.CancelFunc, interval time.Duration, callback func(context.Context) error, release func()) pollRun {
	t.Helper()
	result := make(chan error, 1)
	finished := make(chan struct{})
	t.Cleanup(func() {
		if release != nil {
			release()
		}
		cancel()
		waitStopped(t, "Poll during cleanup", finished)
	})
	go func() {
		defer close(finished)
		result <- poller.Poll(ctx, interval, callback)
	}()
	return pollRun{result: result, finished: finished}
}

func receive[T any](t *testing.T, operation string, events <-chan T) T {
	t.Helper()
	timer := time.NewTimer(testDeadline)
	defer timer.Stop()
	select {
	case value := <-events:
		return value
	case <-timer.C:
		t.Fatalf("timed out waiting for %s", operation)
		var zero T
		return zero
	}
}

func waitStopped(t *testing.T, operation string, finished <-chan struct{}) bool {
	t.Helper()
	timer := time.NewTimer(testDeadline)
	defer timer.Stop()
	select {
	case <-finished:
		return true
	case <-timer.C:
		t.Errorf("timed out waiting for %s", operation)
		return false
	}
}

func gate() (<-chan struct{}, func()) {
	ch := make(chan struct{})
	var once sync.Once
	return ch, func() { once.Do(func() { close(ch) }) }
}

func TestCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan context.Context, 1)
	var calls atomic.Int32
	run := startPoll(t, ctx, cancel, time.Hour, func(callbackCtx context.Context) error {
		calls.Add(1)
		started <- callbackCtx
		return nil
	}, nil)
	if got := receive(t, "first callback", started); got != ctx {
		t.Fatal("callback did not receive the caller context")
	}
	cancel()
	if err := receive(t, "Poll cancellation", run.result); err != nil {
		t.Fatalf("Poll after cancellation = %v, want nil", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("callback count = %d, want 1", got)
	}
}

func TestCallbackError(t *testing.T) {
	t.Parallel()
	want := errors.New("poll failed")
	err := poller.Poll(context.Background(), time.Hour, func(context.Context) error { return want })
	if err != want {
		t.Fatalf("Poll = %v, want unchanged callback error %v", err, want)
	}
}

func TestCallbackErrorDuringCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	want := errors.New("callback failed while canceled")
	err := poller.Poll(ctx, time.Hour, func(callbackCtx context.Context) error {
		if callbackCtx != ctx {
			t.Error("callback did not receive the caller context")
		}
		cancel()
		return want
	})
	if err != want {
		t.Fatalf("Poll = %v, want unchanged callback error %v", err, want)
	}
}

func TestAlreadyCanceled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := poller.Poll(ctx, time.Hour, func(context.Context) error {
		calls++
		return errors.New("unexpected callback")
	})
	if err != nil || calls != 0 {
		t.Fatalf("Poll = %v, callback count = %d; want nil and 0", err, calls)
	}
}

func TestInvalidInterval(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		interval time.Duration
	}{
		{name: "zero", interval: 0},
		{name: "negative", interval: -time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			calls := 0
			err := poller.Poll(context.Background(), test.interval, func(context.Context) error {
				calls++
				return errors.New("unexpected callback")
			})
			if err == nil || calls != 0 {
				t.Fatalf("Poll = %v, callback count = %d; want error and 0", err, calls)
			}
		})
	}
}

func TestRecurrence(t *testing.T) {
	t.Parallel()
	const interval = 25 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	released, release := gate()
	type event struct {
		call int32
		at   time.Time
	}
	starts := make(chan event, 4)
	completions := make(chan event, 4)
	var calls, active atomic.Int32
	var overlap atomic.Bool
	want := errors.New("three cycles complete")
	run := startPoll(t, ctx, cancel, interval, func(callbackCtx context.Context) error {
		if active.Add(1) != 1 {
			overlap.Store(true)
		}
		call := calls.Add(1)
		defer func() {
			active.Add(-1)
			// This is the final callback action, after its work and cleanup.
			completions <- event{call: call, at: time.Now()}
		}()
		starts <- event{call: call, at: time.Now()}
		if call == 1 {
			select {
			case <-released:
			case <-callbackCtx.Done():
				return nil
			}
		}
		if call >= 3 {
			return want
		}
		return nil
	}, release)
	first := receive(t, "first callback start", starts)
	if first.call != 1 {
		t.Fatalf("first callback = %d, want 1", first.call)
	}
	// Hold a known active callback past a full interval to expose fixed-rate
	// scheduling and overlap. This timer controls the scenario, not readiness.
	hold := time.NewTimer(2 * interval)
	defer hold.Stop()
	select {
	case extra := <-starts:
		t.Fatalf("callback %d started before callback 1 finished", extra.call)
	case <-hold.C:
	}
	release()
	for call := int32(2); call <= 3; call++ {
		previous := receive(t, "preceding callback completion", completions)
		next := receive(t, "recurring callback start", starts)
		if previous.call != call-1 || next.call != call {
			t.Fatalf("completion/start = %d/%d, want %d/%d", previous.call, next.call, call-1, call)
		}
		// The timestamp precedes the actual return by only the final channel
		// send, so it is a conservative lower bound. Late starts are allowed.
		if elapsed := next.at.Sub(previous.at); elapsed < interval {
			t.Fatalf("callback %d started %v after completion, want at least %v", call, elapsed, interval)
		}
	}
	if err := receive(t, "Poll after recurrence", run.result); err != want {
		t.Fatalf("Poll = %v, want %v", err, want)
	}
	if overlap.Load() || active.Load() != 0 || calls.Load() != 3 {
		t.Fatalf("overlap = %v, active = %d, calls = %d; want false, 0, 3", overlap.Load(), active.Load(), calls.Load())
	}
}

func TestFixtureTeardown(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "callback.txt")
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	released, release := gate()
	callbackFinished := make(chan struct{})
	pollFinished := make(chan struct{})
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	// Registered before resource cleanup so this observes the actual teardown.
	t.Cleanup(func() {
		if _, err := file.Stat(); !errors.Is(err, os.ErrClosed) {
			t.Errorf("fixture after teardown: %v, want closed file", err)
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("read fixture after teardown: %v", err)
			return
		}
		if string(contents) != "active\ncleanup\ncaller\n" {
			t.Errorf("fixture contents = %q, want callback work, cleanup, and caller work", contents)
		}
	})
	t.Cleanup(func() {
		release()
		cancel()
		callbackStopped := waitStopped(t, "callback cleanup", callbackFinished)
		pollStopped := waitStopped(t, "Poll teardown", pollFinished)
		if !callbackStopped || !pollStopped {
			// A timeout cannot stop a borrower; retain its resource on failure.
			return
		}
		if err := file.Close(); err != nil {
			t.Errorf("close caller fixture: %v", err)
		}
	})
	started := make(chan error, 1)
	cleanupStarted := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		defer close(pollFinished)
		result <- poller.Poll(ctx, time.Hour, func(callbackCtx context.Context) error {
			defer close(callbackFinished)
			_, err := file.WriteString("active\n")
			started <- err
			if err != nil {
				return err
			}
			<-callbackCtx.Done()
			close(cleanupStarted)
			<-released
			_, err = file.WriteString("cleanup\n")
			return err
		})
	}()
	if err := receive(t, "fixture callback startup", started); err != nil {
		t.Fatalf("callback fixture use: %v", err)
	}
	cancel()
	receive(t, "callback entering cleanup", cleanupStarted)
	// Observe return only after the callback has entered its blocked cleanup.
	observation := time.NewTimer(25 * time.Millisecond)
	defer observation.Stop()
	select {
	case err := <-result:
		t.Fatalf("Poll returned %v while callback cleanup was blocked", err)
	case <-observation.C:
	}
	release()
	if err := receive(t, "Poll after callback cleanup", result); err != nil {
		t.Fatalf("Poll after cancellation = %v, want nil", err)
	}
	receive(t, "callback finished", callbackFinished)
	if _, err := file.Stat(); err != nil {
		t.Fatalf("caller fixture unavailable after Poll: %v", err)
	}
	if _, err := file.WriteString("caller\n"); err != nil {
		t.Fatalf("caller fixture write after Poll: %v", err)
	}
}

func TestIndependentInvocations(t *testing.T) {
	t.Parallel()
	t.Run("callback_error", func(t *testing.T) {
		t.Parallel()
		slowCtx, cancelSlow := context.WithCancel(context.Background())
		slowStarted := make(chan context.Context, 1)
		slow := startPoll(t, slowCtx, cancelSlow, time.Hour, func(ctx context.Context) error {
			slowStarted <- ctx
			<-ctx.Done()
			return nil
		}, nil)
		if got := receive(t, "slow invocation callback", slowStarted); got != slowCtx {
			t.Fatal("slow callback received another invocation's context")
		}
		fastCtx, cancelFast := context.WithCancel(context.Background())
		var calls atomic.Int32
		want := errors.New("fast invocation complete")
		fast := startPoll(t, fastCtx, cancelFast, time.Millisecond, func(ctx context.Context) error {
			if ctx != fastCtx {
				return errors.New("fast callback received another invocation's context")
			}
			if calls.Add(1) >= 3 {
				return want
			}
			return nil
		}, nil)
		if err := receive(t, "fast invocation recurrence", fast.result); err != want {
			t.Fatalf("fast Poll = %v, want %v", err, want)
		}
		if got := calls.Load(); got != 3 {
			t.Fatalf("fast callback count = %d, want 3", got)
		}
		select {
		case err := <-slow.result:
			t.Fatalf("slow Poll returned %v before its own cancellation", err)
		default:
		}
		cancelSlow()
		if err := receive(t, "slow invocation cancellation", slow.result); err != nil {
			t.Fatalf("slow Poll = %v, want nil", err)
		}
	})
	t.Run("cancellation", func(t *testing.T) {
		t.Parallel()
		firstCtx, cancelFirst := context.WithCancel(context.Background())
		firstStarted := make(chan struct{})
		first := startPoll(t, firstCtx, cancelFirst, time.Hour, func(ctx context.Context) error {
			close(firstStarted)
			<-ctx.Done()
			return nil
		}, nil)
		receive(t, "first invocation callback", firstStarted)
		secondCtx, cancelSecond := context.WithCancel(context.Background())
		secondStarted := make(chan struct{})
		released, release := gate()
		want := errors.New("second invocation complete")
		second := startPoll(t, secondCtx, cancelSecond, time.Hour, func(ctx context.Context) error {
			close(secondStarted)
			select {
			case <-released:
				return want
			case <-ctx.Done():
				return ctx.Err()
			}
		}, release)
		receive(t, "second invocation callback", secondStarted)
		cancelFirst()
		if err := receive(t, "first invocation cancellation", first.result); err != nil {
			t.Fatalf("first Poll = %v, want nil", err)
		}
		if err := secondCtx.Err(); err != nil {
			t.Fatalf("first cancellation canceled second context: %v", err)
		}
		select {
		case err := <-second.result:
			t.Fatalf("second Poll returned %v before its own release", err)
		default:
		}
		release()
		if err := receive(t, "second invocation completion", second.result); err != want {
			t.Fatalf("second Poll = %v, want %v", err, want)
		}
	})
}
