package pipeline

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type recoveryCause struct{ details []string }

func (recoveryCause) Error() string { return "caller stopped" }

func recoveryAwait(t *testing.T, event <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-event:
	case <-time.After(time.Second):
		t.Fatalf("missing %s", what)
	}
}

func TestRecoveryEarlyConsumerCompletionJoinsProducer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopping := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	done := make(chan error, 1)
	accepted := 0
	go func() {
		done <- runPipeline(ctx, 1, func(ctx context.Context, out chan<- int) error {
			for value := 1; ; value++ {
				select {
				case out <- value:
				case <-ctx.Done():
					close(stopping)
					<-release
					return ctx.Err()
				}
			}
		}, func(ctx context.Context, in <-chan int) error {
			select {
			case accepted = <-in:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	recoveryAwait(t, stopping, "producer stopping after early consumer success")
	select {
	case err := <-done:
		t.Fatalf("returned before producer cleanup: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	once.Do(func() { close(release) })
	select {
	case err := <-done:
		if err != nil || accepted != 1 {
			t.Fatalf("accepted=%d err=%v", accepted, err)
		}
	case <-time.After(time.Second):
		t.Fatal("pipeline did not join")
	}
}

func TestRecoveryIndependentCancellationClassFailures(t *testing.T) {
	for _, producerErr := range []error{context.DeadlineExceeded, errors.Join(context.Canceled, errors.New("producer independently failed"))} {
		consumerErr := errors.Join(context.DeadlineExceeded, errors.New("consumer independently failed"))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		consumerStarted := make(chan struct{})
		err := runPipeline(ctx, 1, func(ctx context.Context, _ chan<- int) error {
			select {
			case <-consumerStarted:
				return producerErr
			case <-ctx.Done():
				return ctx.Err()
			}
		}, func(ctx context.Context, _ <-chan int) error {
			close(consumerStarted)
			<-ctx.Done()
			return consumerErr
		})
		cancel()
		if !errors.Is(err, producerErr) || !errors.Is(err, consumerErr) {
			t.Fatalf("independent outcomes lost: %v", err)
		}
	}
}

func TestRecoveryCallerCauseAndBothJoins(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	producerStarted := make(chan struct{})
	consumerStarted := make(chan struct{})
	producerStopped := make(chan struct{})
	consumerStopped := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	defer once.Do(func() { close(release) })
	done := make(chan error, 1)
	go func() {
		done <- runPipeline(ctx, 1, func(ctx context.Context, _ chan<- int) error {
			close(producerStarted)
			<-ctx.Done()
			close(producerStopped)
			<-release
			return ctx.Err()
		}, func(ctx context.Context, _ <-chan int) error {
			close(consumerStarted)
			<-ctx.Done()
			close(consumerStopped)
			<-release
			return ctx.Err()
		})
	}()
	recoveryAwait(t, producerStarted, "producer start")
	recoveryAwait(t, consumerStarted, "consumer start")
	cancel(recoveryCause{details: []string{"metadata"}})
	recoveryAwait(t, producerStopped, "producer stop")
	recoveryAwait(t, consumerStopped, "consumer stop")
	select {
	case err := <-done:
		t.Fatalf("returned before callback cleanup: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	once.Do(func() { close(release) })
	select {
	case err := <-done:
		var cause recoveryCause
		if !errors.Is(err, context.Canceled) || !errors.As(err, &cause) {
			t.Fatalf("caller outcome lost: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("pipeline did not finish")
	}
}

func TestRecoveryCanceledEntry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := runPipeline(ctx, 1, func(context.Context, chan<- int) error { t.Error("producer admitted"); return nil }, func(context.Context, <-chan int) error { t.Error("consumer admitted"); return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("entry cancellation lost: %v", err)
	}
}
