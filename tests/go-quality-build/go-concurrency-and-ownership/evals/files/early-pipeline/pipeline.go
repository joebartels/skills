package pipeline

import "context"

func runPipeline(ctx context.Context, capacity int, produce func(context.Context, chan<- int) error, consume func(context.Context, <-chan int) error) error {
	values := make(chan int, capacity)
	if err := produce(ctx, values); err != nil {
		return err
	}
	close(values)
	return consume(ctx, values)
}
