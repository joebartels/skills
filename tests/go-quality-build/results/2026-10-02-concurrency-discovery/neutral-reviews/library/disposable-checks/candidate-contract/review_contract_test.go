package totals_test

import (
	"sync"
	"testing"

	"example.com/library-shared-state"
)

func TestReviewZeroValueAndSnapshotOwnership(t *testing.T) {
	var got totals.Totals
	if first := got.Snapshot(); first != (totals.Snapshot{}) {
		t.Fatalf("zero snapshot=%+v", first)
	}
	got.Add(3)
	before := got.Snapshot()
	got.Add(-2)
	if before != (totals.Snapshot{Count: 1, Sum: 3}) {
		t.Fatalf("previous snapshot changed: %+v", before)
	}
	before.Count = -100
	before.Sum = -100
	if current := got.Snapshot(); current != (totals.Snapshot{Count: 2, Sum: 1}) {
		t.Fatalf("snapshot mutation reached Totals: %+v", current)
	}
}

func TestReviewConcurrentReadersAndIndependentInstances(t *testing.T) {
	const updates = 5000
	var first, second totals.Totals
	var wg sync.WaitGroup
	start := make(chan struct{})
	errors := make(chan totals.Snapshot, 4)
	for _, instance := range []*totals.Totals{&first, &second} {
		instance := instance
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < updates; i++ {
				instance.Add(-3)
			}
		}()
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				var previous int64
				for i := 0; i < updates; i++ {
					current := instance.Snapshot()
					if current.Sum != -3*current.Count || current.Count < previous {
						errors <- current
						return
					}
					previous = current.Count
				}
			}()
		}
	}
	close(start)
	wg.Wait()
	close(errors)
	for snapshot := range errors {
		t.Errorf("incoherent or regressing snapshot: %+v", snapshot)
	}
	for _, instance := range []*totals.Totals{&first, &second} {
		if current := instance.Snapshot(); current != (totals.Snapshot{Count: updates, Sum: -3 * updates}) {
			t.Errorf("instance snapshot=%+v", current)
		}
	}
}
