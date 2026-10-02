package example

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestZeroValueAndReturnedScalars(t *testing.T) {
	var state totals
	if count, sum := state.snapshot(); count != 0 || sum != 0 {
		t.Fatalf("zero snapshot = %d, %d", count, sum)
	}
	state.add(5)
	count, sum := state.snapshot()
	state.add(-2)
	if count != 1 || sum != 5 {
		t.Fatalf("prior returned values changed: %d, %d", count, sum)
	}
	if count, sum := state.snapshot(); count != 2 || sum != 3 {
		t.Fatalf("updated snapshot = %d, %d; want 2, 3", count, sum)
	}
}

func TestConcurrentCoherentSnapshots(t *testing.T) {
	const workers, updates, delta = 12, 2000, 7
	var state totals
	var writers sync.WaitGroup
	writers.Add(workers)
	start := make(chan struct{})
	finished := make(chan struct{})
	observed := make(chan error, 1)
	for worker := 0; worker < workers; worker++ {
		go func() {
			defer writers.Done()
			<-start
			for n := 0; n < updates; n++ {
				state.add(delta)
			}
		}()
	}
	go func() {
		writers.Wait()
		close(finished)
	}()
	go func() {
		<-start
		for {
			count, sum := state.snapshot()
			if count < 0 || count > workers*updates || sum != delta*count {
				observed <- fmt.Errorf("incoherent snapshot = %d, %d", count, sum)
				return
			}
			select {
			case <-finished:
				observed <- nil
				return
			default:
			}
		}
	}()
	close(start)
	select {
	case err := <-observed:
		if err != nil {
			t.Error(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("concurrent snapshot observation did not finish")
	}
	select {
	case <-finished:
	case <-time.After(10 * time.Second):
		t.Fatal("writers did not finish")
	}
	if count, sum := state.snapshot(); count != workers*updates || sum != delta*workers*updates {
		t.Fatalf("final snapshot = %d, %d; want %d, %d", count, sum, workers*updates, delta*workers*updates)
	}
}

func TestSeparateInstances(t *testing.T) {
	var first, second totals
	first.add(13)
	second.add(-4)
	if count, sum := first.snapshot(); count != 1 || sum != 13 {
		t.Fatalf("first instance = %d, %d", count, sum)
	}
	if count, sum := second.snapshot(); count != 1 || sum != -4 {
		t.Fatalf("second instance = %d, %d", count, sum)
	}
}
