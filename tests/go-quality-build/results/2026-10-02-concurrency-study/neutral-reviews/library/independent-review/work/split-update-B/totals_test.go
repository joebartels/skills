package totals

import (
	"sync"
	"testing"
)

func TestSequentialTotals(t *testing.T) {
	var got Totals
	got.Add(3)
	got.Add(-2)
	if got.Snapshot() != (Snapshot{Count: 2, Sum: 1}) {
		t.Fatalf("totals=%+v", got.Snapshot())
	}
}

func TestConcurrentAddsAndSnapshots(t *testing.T) {
	var got Totals
	const writers = 8
	const addsPerWriter = 1000
	start := make(chan struct{})
	var wg sync.WaitGroup

	for range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for range addsPerWriter {
				got.Add(1)
			}
		}()
	}

	stopReader := make(chan struct{})
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		<-start
		for {
			select {
			case <-stopReader:
				return
			default:
				if snapshot := got.Snapshot(); snapshot.Count != snapshot.Sum {
					t.Errorf("incoherent snapshot: %+v", snapshot)
					return
				}
			}
		}
	}()

	close(start)
	wg.Wait()
	close(stopReader)
	<-readerDone
	// A final snapshot establishes the exact accepted update count.
	if snapshot := got.Snapshot(); snapshot != (Snapshot{Count: writers * addsPerWriter, Sum: writers * addsPerWriter}) {
		t.Fatalf("final snapshot = %+v", snapshot)
	}
}

func TestSnapshotsAreIndependentAndInstancesAreIsolated(t *testing.T) {
	var first, second Totals
	first.Add(4)
	snapshot := first.Snapshot()
	second.Add(7)
	first.Add(2)

	if snapshot != (Snapshot{Count: 1, Sum: 4}) {
		t.Fatalf("earlier snapshot changed: %+v", snapshot)
	}
	if got := first.Snapshot(); got != (Snapshot{Count: 2, Sum: 6}) {
		t.Fatalf("first totals = %+v", got)
	}
	if got := second.Snapshot(); got != (Snapshot{Count: 1, Sum: 7}) {
		t.Fatalf("second totals = %+v", got)
	}
}
