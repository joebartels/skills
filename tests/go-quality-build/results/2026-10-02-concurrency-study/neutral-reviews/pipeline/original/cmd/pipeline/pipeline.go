package main

import (
	"context"
	"errors"
)

func runPipeline(ctx context.Context, capacity int, produce func(context.Context, chan<- int) error, consume func(context.Context, <-chan int) error) error {
	out := make(chan int, capacity)
	done := make(chan error, 1)
	go func() { defer close(out); done <- produce(ctx, out) }()
	consumerErr := consume(ctx, out)
	return errors.Join(consumerErr, <-done)
}
