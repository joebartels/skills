package sweeper_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/sweeper"
)

func TestHostJoinBeforeRelease(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	cleaning := make(chan struct{})
	permit := make(chan struct{})
	order := make(chan string, 2)
	done := make(chan error, 1)
	defer close(permit)
	go func() {
		done <- sweeper.Run(ctx, time.Millisecond, func(ctx context.Context) error {
			close(started)
			<-ctx.Done()
			close(cleaning)
			select {
			case <-permit:
			case <-time.After(time.Second):
			}
			order <- "callback finished"
			return nil
		}, func() error { order <- "resource released"; return nil })
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("callback did not start")
	}
	cancel()
	select {
	case <-cleaning:
	case <-time.After(2 * time.Second):
		t.Fatal("callback did not receive cancellation")
	}
	// Permit cleanup after startup/cancellation observation. Send avoids double-close cleanup.
	permit <- struct{}{}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("host did not finish")
	}
	var events []string
	for len(events) < 2 {
		select {
		case e := <-order:
			events = append(events, e)
		case <-time.After(2 * time.Second):
			t.Fatal("missing completion/release event")
		}
	}
	if events[0] != "callback finished" || events[1] != "resource released" {
		t.Fatalf("order = %v; host must join before release", events)
	}
}

func TestSuccessfulRecurrence(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	calls := 0
	err := sweeper.Run(ctx, 5*time.Millisecond, func(context.Context) error {
		calls++
		if calls == 2 {
			cancel()
		}
		return nil
	}, func() error { return nil })
	if err != nil || calls != 2 {
		t.Fatalf("Run = %v, successful calls = %d; want 2", err, calls)
	}
}

func TestErrorsAndInvalidInterval(t *testing.T) {
	work, release := errors.New("work"), errors.New("release")
	err := sweeper.Run(context.Background(), time.Second, func(context.Context) error { return work }, func() error { return release })
	if !errors.Is(err, work) || !errors.Is(err, release) {
		t.Errorf("Run lost error identity: %v", err)
	}
	called := false
	err = sweeper.Run(context.Background(), 0, func(context.Context) error { called = true; return nil }, func() error { called = true; return nil })
	if err == nil || called {
		t.Errorf("invalid interval: err=%v called=%v", err, called)
	}
}
