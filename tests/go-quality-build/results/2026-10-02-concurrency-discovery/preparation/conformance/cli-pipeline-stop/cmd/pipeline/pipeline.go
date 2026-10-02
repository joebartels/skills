package main

import (
	"context"
	"errors"
)

type pipelineResult struct {
	err     error
	stopped bool
}

func runPipeline(ctx context.Context, capacity int, produce func(context.Context, chan<- int) error, consume func(context.Context, <-chan int) error) error {
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	output := make(chan int, capacity)
	producerDone := make(chan pipelineResult, 1)
	consumerDone := make(chan pipelineResult, 1)
	go func() {
		producerErr := produce(child, output)
		stopped := child.Err() != nil
		close(output)
		if producerErr != nil {
			cancel()
		}
		producerDone <- pipelineResult{producerErr, stopped}
	}()
	go func() {
		consumerErr := consume(child, output)
		stopped := child.Err() != nil
		cancel()
		consumerDone <- pipelineResult{consumerErr, stopped}
	}()
	producer, consumer := <-producerDone, <-consumerDone
	normalize := func(result pipelineResult) error {
		if result.stopped && (result.err == context.Canceled || result.err == context.DeadlineExceeded) {
			return nil
		}
		return result.err
	}
	return errors.Join(normalize(producer), normalize(consumer), ctx.Err())
}
