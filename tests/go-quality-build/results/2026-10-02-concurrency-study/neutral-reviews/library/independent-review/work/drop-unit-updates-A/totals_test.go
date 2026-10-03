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
		writers       = 8
		addsPerWriter = 1000
	)

	var got Totals
	start := make(chan struct{})
	var writersDone sync.WaitGroup
	writersDone.Add(writers)
	for range writers {
		go func() {
			defer writersDone.Done()
			<-start
			for range addsPerWriter {
				got.Add(1)
			}
		}()
	}

	var snapshotsDone sync.WaitGroup
	snapshotsDone.Add(1)
	go func() {
		defer snapshotsDone.Done()
		<-start
		for {
			snapshot := got.Snapshot()
			if snapshot.Count != snapshot.Sum {
				t.Errorf("incoherent snapshot: %+v", snapshot)
				return
			}
			if snapshot.Count == writers*addsPerWriter {
				return
			}
		}
	}()

	close(start)
	writersDone.Wait()
	snapshotsDone.Wait()
	if want := (Snapshot{Count: writers * addsPerWriter, Sum: writers * addsPerWriter}); got.Snapshot() != want {
		t.Fatalf("final totals=%+v, want %+v", got.Snapshot(), want)
	}
}

func TestTotalsInstancesAreIndependent(t *testing.T) {
	var first, second Totals
	first.Add(4)
	second.Add(2)
	first.Add(-1)

	if want := (Snapshot{Count: 2, Sum: 3}); first.Snapshot() != want {
		t.Fatalf("first totals=%+v, want %+v", first.Snapshot(), want)
	}
	if want := (Snapshot{Count: 1, Sum: 2}); second.Snapshot() != want {
		t.Fatalf("second totals=%+v, want %+v", second.Snapshot(), want)
	}
}
