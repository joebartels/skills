package totals

import "testing"

func TestSequentialTotals(t *testing.T) {
	var got Totals
	got.Add(3)
	got.Add(-2)
	if got.Snapshot() != (Snapshot{Count: 2, Sum: 1}) {
		t.Fatalf("totals=%+v", got.Snapshot())
	}
}
