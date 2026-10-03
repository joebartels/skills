package totals

import (
	"runtime"
	"sync"
	"testing"
)

func TestControllerCoherentSnapshots(t *testing.T) {
	var totals Totals
	var wg sync.WaitGroup
	start := make(chan struct{})
	done := make(chan struct{})
	const writers = 4
	const each = 10000
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for n := 0; n < each; n++ {
				totals.Add(1)
				if n%17 == 0 {
					runtime.Gosched()
				}
			}
		}()
	}
	go func() { wg.Wait(); close(done) }()
	close(start)
	bad := false
observe:
	for {
		select {
		case <-done:
			break observe
		default:
			s := totals.Snapshot()
			if s.Count != s.Sum {
				bad = true
				break observe
			}
			runtime.Gosched()
		}
	}
	<-done
	if bad {
		t.Error("incoherent Count/Sum publication")
	}
	if got := totals.Snapshot(); got != (Snapshot{writers * each, writers * each}) {
		t.Fatalf("final=%+v", got)
	}
	var mixed Totals
	wg = sync.WaitGroup{}
	for delta := int64(-2); delta <= 2; delta++ {
		wg.Add(1)
		go func(d int64) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				mixed.Add(d)
			}
		}(delta)
	}
	wg.Wait()
	if got := mixed.Snapshot(); got != (Snapshot{1000, 0}) {
		t.Fatalf("mixed=%+v", got)
	}
}
func TestControllerIndependentTotals(t *testing.T) {
	var a, b Totals
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for n := 0; n < 100; n++ {
				a.Add(2)
			}
		}()
		go func() {
			defer wg.Done()
			for n := 0; n < 150; n++ {
				b.Add(-3)
			}
		}()
	}
	wg.Wait()
	if a.Snapshot() != (Snapshot{200, 400}) || b.Snapshot() != (Snapshot{300, -900}) {
		t.Fatalf("a=%+v b=%+v", a.Snapshot(), b.Snapshot())
	}
	copy := a.Snapshot()
	copy.Sum = 999
	if a.Snapshot().Sum != 400 {
		t.Fatal("snapshot aliases state")
	}
}
