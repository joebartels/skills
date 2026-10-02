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

func TestConcurrentAddAndSnapshot(t *testing.T) {
	const (
		writers = 8
		updates = 2000
	)

	var got Totals
	var wg sync.WaitGroup
	start := make(chan struct{})
	snapshots := make(chan Snapshot, writers)

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < updates; j++ {
				got.Add(1)
				if j%10 == 0 {
					snapshot := got.Snapshot()
					if snapshot.Count != snapshot.Sum {
						snapshots <- snapshot
						return
					}
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(snapshots)
	for snapshot := range snapshots {
		t.Errorf("incoherent snapshot: %+v", snapshot)
	}

	want := int64(writers * updates)
	if snapshot := got.Snapshot(); snapshot != (Snapshot{Count: want, Sum: want}) {
		t.Fatalf("final snapshot=%+v, want count and sum %d", snapshot, want)
	}
}

func TestTotalsInstancesAreIndependent(t *testing.T) {
	var first, second Totals
	first.Add(4)
	second.Add(9)
	first.Add(-1)

	if got, want := first.Snapshot(), (Snapshot{Count: 2, Sum: 3}); got != want {
		t.Errorf("first snapshot=%+v, want %+v", got, want)
	}
	if got, want := second.Snapshot(), (Snapshot{Count: 1, Sum: 9}); got != want {
		t.Errorf("second snapshot=%+v, want %+v", got, want)
	}
}
