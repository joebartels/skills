package main

import (
	"context"
	"errors"
)

func runPipeline(ctx context.Context, capacity int, produce func(context.Context, chan<- int) error, consume func(context.Context, <-chan int) error) error {
	pipelineCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := make(chan int, capacity)
	type result struct {
		producer bool
		err      error
	}
	results := make(chan result, 2)

	go func() {
		err := produce(pipelineCtx, out)
		close(out)
		results <- result{producer: true, err: err}
	}()
	go func() {
		err := consume(pipelineCtx, out)
		results <- result{err: err}
	}()

	first := <-results
	// A successful producer has closed the stream, so let the consumer drain it.
	// Every other first completion coordinates shutdown of the other callback.
	if !first.producer || first.err != nil || ctx.Err() != nil {
		cancel()
	}
	second := <-results

	var producerErr, consumerErr error
	if first.producer {
		producerErr, consumerErr = first.err, second.err
	} else {
		consumerErr, producerErr = first.err, second.err
	}
	return errors.Join(ctx.Err(), suppressStopError(ctx, producerErr), suppressStopError(ctx, consumerErr))
}

func suppressStopError(parent context.Context, err error) error {
	if parent.Err() == nil && (err == context.Canceled || err == context.DeadlineExceeded) {
		return nil
	}
	return err
}
