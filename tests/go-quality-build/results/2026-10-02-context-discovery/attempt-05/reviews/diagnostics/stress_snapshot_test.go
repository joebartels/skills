package stages

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestReviewCancellationSnapshot(t *testing.T) {
	op := errors.New("operation failure")
	cause := errors.New("custom cancellation")
	for i := 0; i < 100000; i++ {
		parent, cancel := context.WithCancelCause(context.Background())
		ctx, release := context.WithTimeout(parent, time.Minute)
		var started sync.WaitGroup
		started.Add(1)
		go func() { started.Done(); cancel(cause) }()
		started.Wait()
		if i%2 == 0 { runtime.Gosched() }
		err := stageError(op, ctx)
		cancel(cause)
		release()
		if errors.Is(err, cause) && !errors.Is(err, context.Canceled) {
			t.Fatalf("iteration %d: cancellation cause retained without cancellation classification: %v", i, err)
		}
	}
}

type reviewTransitionContext struct {
	context.Context
	cancel func()
}

func (ctx reviewTransitionContext) Err() error {
	before := ctx.Context.Err()
	ctx.cancel()
	return before
}

func TestReviewErrorRepresentationDuringCancellationTransition(t *testing.T) {
	op := errors.New("independent failure")
	cause := errors.New("custom reason")
	base, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	ctx := reviewTransitionContext{Context: base, cancel: func() { cancel(cause) }}
	err := stageError(op, ctx)
	if !errors.Is(err, op) { t.Fatal("lost independent operation error") }
	if errors.Is(err, cause) && !errors.Is(err, context.Canceled) {
		t.Fatalf("custom cancellation cause retained without classification: %v", err)
	}
}

func TestReviewPublicFailureCancellationSnapshot(t *testing.T) {
	op := errors.New("transport failure")
	cause := errors.New("custom cancellation")
	for i := 0; i < 300000; i++ {
		ctx, cancel := context.WithCancelCause(context.Background())
		var done sync.WaitGroup
		done.Add(1)
		var started sync.WaitGroup
		started.Add(1)
		client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
			go func() { defer done.Done(); started.Done(); cancel(cause) }()
			started.Wait()
			return nil, op
		})}
		_, err := FetchAll(ctx, client, []string{"http://example.test/"}, time.Minute, time.Minute)
		done.Wait()
		cancel(cause)
		if errors.Is(err, cause) && !errors.Is(err, context.Canceled) {
			t.Fatalf("iteration %d: public failure retained custom cause without cancellation classification: %v", i, err)
		}
	}
}
