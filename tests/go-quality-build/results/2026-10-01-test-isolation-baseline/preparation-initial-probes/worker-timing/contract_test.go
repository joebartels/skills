package poller_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"example.com/poller"
)

func TestCompletionRelativeRecurrence(t *testing.T) {
	const interval = 40 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stop := errors.New("observed recurrence")
	started := make(chan time.Time, 2)
	finishFirst := make(chan struct{})
	var once sync.Once
	var calls atomic.Int32
	done := make(chan error, 1)
	go func() {
		done <- poller.Poll(ctx, interval, func(context.Context) error {
			started <- time.Now()
			if calls.Add(1) == 1 {
				select {
				case <-finishFirst:
					return nil
				case <-ctx.Done():
					return stop
				}
			}
			return stop
		})
	}()
	t.Cleanup(func() {
		cancel()
		once.Do(func() { close(finishFirst) })
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("Poll cleanup did not join")
		}
	})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first callback did not start")
	}
	// Deliberate work duration creates a stale-ticker risk; readiness is observed.
	<-time.After(2 * interval)
	finished := time.Now()
	once.Do(func() { close(finishFirst) })
	select {
	case next := <-started:
		if next.Sub(finished) < interval {
			t.Fatalf("next callback after %v, want >= %v", next.Sub(finished), interval)
		}
	case <-time.After(time.Second):
		t.Fatal("second callback did not start")
	}
	select {
	case err := <-done:
		if !errors.Is(err, stop) {
			t.Fatalf("Poll = %v, want callback cause", err)
		}
		// Make terminal outcome available to cleanup without another worker.
		done <- err
	case <-time.After(time.Second):
		t.Fatal("Poll did not report recurrence outcome")
	}
}

func TestTeardownOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	cleaning := make(chan struct{})
	allowCleanup := make(chan struct{})
	returned := make(chan error, 1)
	var once sync.Once
	var cleaned, resourceClosed atomic.Bool
	go func() {
		returned <- poller.Poll(ctx, time.Hour, func(callbackCtx context.Context) error {
			close(started)
			<-callbackCtx.Done()
			close(cleaning)
			<-allowCleanup
			if resourceClosed.Load() {
				return errors.New("resource closed during callback cleanup")
			}
			cleaned.Store(true)
			return nil
		})
	}()
	t.Cleanup(func() {
		cancel()
		once.Do(func() { close(allowCleanup) })
		select {
		case <-returned:
		case <-time.After(time.Second):
			t.Error("teardown did not join Poll")
		}
		resourceClosed.Store(true)
	})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("callback did not start")
	}
	cancel()
	select {
	case <-cleaning:
	case <-time.After(time.Second):
		t.Fatal("caller cancellation did not reach callback cleanup")
	}
	once.Do(func() { close(allowCleanup) })
	select {
	case err := <-returned:
		if err != nil || !cleaned.Load() {
			t.Fatalf("joined cleanup=%v, result=%v", cleaned.Load(), err)
		}
		returned <- err
	case <-time.After(time.Second):
		t.Fatal("Poll did not join cleanup")
	}
}
