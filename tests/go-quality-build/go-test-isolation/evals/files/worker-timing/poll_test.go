package poller_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/poller"
)

func TestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- poller.Poll(ctx, time.Hour, func(context.Context) error { return nil }) }()
	time.Sleep(5 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("Poll did not stop after cancellation")
	}
}

func TestCallbackError(t *testing.T) {
	want := errors.New("poll failed")
	err := poller.Poll(context.Background(), time.Hour, func(context.Context) error { return want })
	if !errors.Is(err, want) {
		t.Fatalf("Poll = %v, want callback cause", err)
	}
}
