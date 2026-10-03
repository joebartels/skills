package pipeline

import (
	"context"
	"testing"
)

func TestOneValue(t *testing.T) {
	got := 0
	err := runPipeline(context.Background(), 1, func(ctx context.Context, out chan<- int) error { out <- 7; return nil }, func(ctx context.Context, in <-chan int) error {
		for value := range in {
			got += value
		}
		return nil
	})
	if err != nil || got != 7 {
		t.Fatalf("got=%d err=%v", got, err)
	}
}
