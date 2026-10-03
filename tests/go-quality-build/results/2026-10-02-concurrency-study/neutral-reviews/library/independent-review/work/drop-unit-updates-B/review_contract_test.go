package totals_test

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	totals "example.com/library-shared-state"
)

// Compile the exact existing public method types as an external consumer.
var _ func(*totals.Totals, int64) = (*totals.Totals).Add
var _ func(*totals.Totals) totals.Snapshot = (*totals.Totals).Snapshot

func TestReviewZeroMixedAndValueOwnership(t *testing.T) {
	var first, second totals.Totals
	if got := first.Snapshot(); got != (totals.Snapshot{}) {
		t.Fatalf("zero value: got %+v, want zero", got)
	}
	for _, delta := range []int64{0, 7, -4, 0, -3} {
		first.Add(delta)
	}
	old := first.Snapshot()
	if old != (totals.Snapshot{Count: 5, Sum: 0}) {
		t.Fatalf("mixed and zero deltas: got %+v, want {5 0}", old)
	}
	first.Add(9)
	if old != (totals.Snapshot{Count: 5, Sum: 0}) {
		t.Fatalf("older snapshot changed: %+v", old)
	}
	old.Count, old.Sum = -999, -999
	if got := first.Snapshot(); got != (totals.Snapshot{Count: 6, Sum: 9}) {
		t.Fatalf("caller modification affected totals: %+v", got)
	}
	second.Add(-6)
	if got := second.Snapshot(); got != (totals.Snapshot{Count: 1, Sum: -6}) {
		t.Fatalf("second instance: %+v", got)
	}
	if got := first.Snapshot(); got != (totals.Snapshot{Count: 6, Sum: 9}) {
		t.Fatalf("second update affected first: %+v", got)
	}
}

func TestReviewConcurrentCheckpoint(t *testing.T) {
	const writers, each = 4, 1000
	var got totals.Totals
	start := make(chan struct{})
	resume := make(chan struct{})
	checkpoint := make(chan struct{}, writers)
	stopReader := make(chan struct{})
	readerReady := make(chan struct{})
	readerDone := make(chan struct{})
	bad := make(chan string, 1)
	var wg sync.WaitGroup

	go func() {
		defer close(readerDone)
		initial := got.Snapshot()
		if initial != (totals.Snapshot{}) {
			bad <- fmt.Sprintf("initial reader snapshot: %+v", initial)
		}
		close(readerReady)
		for {
			select {
			case <-stopReader:
				return
			default:
				observation := got.Snapshot()
				if observation.Count != observation.Sum {
					select {
					case bad <- fmt.Sprintf("incoherent snapshot: %+v", observation):
					default:
					}
					return
				}
				runtime.Gosched()
			}
		}
	}()

	wait := func(ch <-chan struct{}, what string) {
		t.Helper()
		select {
		case <-ch:
		case <-time.After(3 * time.Second):
			t.Fatalf("timed out waiting for %s", what)
		}
	}
	wait(readerReady, "initial reader observation")
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			got.Add(1)
			checkpoint <- struct{}{}
			<-resume
			for n := 1; n < each; n++ {
				got.Add(1)
				if n%17 == 0 {
					runtime.Gosched()
				}
			}
		}()
	}
	close(start)
	for i := 0; i < writers; i++ {
		wait(checkpoint, "writer checkpoint")
	}
	// Every writer is alive and explicitly parked. This intermediate snapshot
	// cannot be skipped by a scheduler that finishes all writers first.
	partial := got.Snapshot()
	close(resume)
	writersDone := make(chan struct{})
	go func() { wg.Wait(); close(writersDone) }()
	wait(writersDone, "writer completion")
	close(stopReader)
	wait(readerDone, "reader completion")
	if partial != (totals.Snapshot{Count: writers, Sum: writers}) {
		t.Errorf("checkpoint snapshot: got %+v, want {%d %d}", partial, writers, writers)
	}
	select {
	case message := <-bad:
		t.Error(message)
	default:
	}
	if final := got.Snapshot(); final != (totals.Snapshot{Count: writers * each, Sum: writers * each}) {
		t.Fatalf("final snapshot: got %+v, want {%d %d}", final, writers*each, writers*each)
	}
}

func TestReviewConstantResourceUse(t *testing.T) {
	var got totals.Totals
	allocations := testing.AllocsPerRun(100, func() {
		got.Add(1)
		_ = got.Snapshot()
	})
	if allocations != 0 {
		t.Fatalf("Add/Snapshot allocated %g times per call pair", allocations)
	}
}
