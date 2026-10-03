package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestFinitePipeline(t *testing.T) {
	var got []int
	err := runPipeline(context.Background(), 1, func(ctx context.Context, out chan<- int) error {
		for _, v := range []int{1, 2} {
			select {
			case out <- v:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}, func(ctx context.Context, in <-chan int) error {
		for v := range in {
			got = append(got, v)
		}
		return nil
	})
	if err != nil || !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestConsumerEarlyCompletionCancelsAndJoinsProducer(t *testing.T) {
	started := make(chan struct{})
	cancelled := make(chan struct{})
	cleanup := make(chan struct{})
	joined := make(chan struct{})
	err := make(chan error, 1)
	go func() {
		err <- runPipeline(context.Background(), 1, func(ctx context.Context, out chan<- int) error {
			defer close(joined)
			out <- 1
			close(started)
			select {
			case out <- 2: // capacity fills; the next send blocks until cancellation.
			case <-ctx.Done():
				close(cancelled)
				<-cleanup
				return ctx.Err()
			}
			select {
			case out <- 3:
				return nil
			case <-ctx.Done():
				close(cancelled)
				<-cleanup
				return ctx.Err()
			}
		}, func(ctx context.Context, in <-chan int) error {
			<-started
			<-in
			return nil // take-one behavior
		})
	}()

	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("producer did not observe cancellation")
	}
	select {
	case <-joined:
		t.Fatal("producer returned before its cleanup gate opened")
	default:
	}
	close(cleanup)
	select {
	case got := <-err:
		if got != nil {
			t.Fatalf("runPipeline error = %v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("runPipeline did not join producer")
	}
	select {
	case <-joined:
	default:
		t.Fatal("producer was not joined before return")
	}
}

func TestProducerFailureStopsConsumerAndPreservesBothErrors(t *testing.T) {
	producerFailure := errors.Join(context.Canceled, errors.New("producer failed"))
	consumerFailure := errors.Join(context.Canceled, errors.New("consumer failed"))
	consumerStarted := make(chan struct{})
	err := runPipeline(context.Background(), 1, func(ctx context.Context, out chan<- int) error {
		<-consumerStarted
		return producerFailure
	}, func(ctx context.Context, in <-chan int) error {
		close(consumerStarted)
		<-ctx.Done()
		return fmt.Errorf("consumer cleanup: %w", consumerFailure)
	})
	if !errors.Is(err, producerFailure) || !errors.Is(err, consumerFailure) {
		t.Fatalf("error = %v, want both independent failures", err)
	}
}

func TestCoordinatedCancellationIsSuppressedButCallerCancellationIsNot(t *testing.T) {
	err := runPipeline(context.Background(), 1, func(ctx context.Context, out chan<- int) error {
		<-ctx.Done()
		return ctx.Err()
	}, func(context.Context, <-chan int) error { return nil })
	if err != nil {
		t.Fatalf("coordinated stop error = %v, want nil", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = runPipeline(ctx, 1, func(ctx context.Context, out chan<- int) error {
		<-ctx.Done()
		return ctx.Err()
	}, func(ctx context.Context, in <-chan int) error {
		<-ctx.Done()
		return ctx.Err()
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("caller cancellation error = %v, want context.Canceled", err)
	}
}
