package main

import (
	"context"
	"errors"
)

func runPipeline(ctx context.Context, capacity int, produce func(context.Context, chan<- int) error, consume func(context.Context, <-chan int) error) error {
	out := make(chan int, capacity)
	workCtx, cancel := context.WithCancel(ctx)
	type result struct {
		producer        bool
		stoppedAtReturn bool
		err             error
	}
	results := make(chan result, 2)
	go func() {
		err := produce(workCtx, out)
		results <- result{producer: true, stoppedAtReturn: workCtx.Err() != nil, err: err}
		close(out)
	}()
	go func() {
		err := consume(workCtx, out)
		results <- result{stoppedAtReturn: workCtx.Err() != nil, err: err}
	}()

	var producerErr, consumerErr error
	producerDone, consumerDone := false, false
	producerStopped, consumerStopped := false, false
	callerDone := ctx.Done()
	for !producerDone || !consumerDone {
		select {
		case <-callerDone:
			callerDone = nil
			cancel()
		case r := <-results:
			if r.producer {
				producerDone, producerErr = true, r.err
				producerStopped = r.stoppedAtReturn
				if r.err != nil {
					cancel()
				}
			} else {
				consumerDone, consumerErr = true, r.err
				consumerStopped = r.stoppedAtReturn
				cancel()
			}
		}
	}
	cancel()

	var errs []error
	for i, err := range []error{producerErr, consumerErr} {
		if err == nil {
			continue
		}
		suppress := (i == 0 && producerStopped) || (i == 1 && consumerStopped)
		if ctx.Err() == nil && suppress && (err == context.Canceled || err == context.DeadlineExceeded) {
			continue
		}
		errs = append(errs, err)
	}
	if err := ctx.Err(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
