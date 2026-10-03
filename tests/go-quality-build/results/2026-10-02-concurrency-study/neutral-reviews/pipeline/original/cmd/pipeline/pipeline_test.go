package main

import (
	"context"
	"reflect"
	"testing"
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
