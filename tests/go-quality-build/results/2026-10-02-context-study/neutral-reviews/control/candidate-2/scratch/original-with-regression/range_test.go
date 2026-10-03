package rangecheck

import "testing"

func TestInteriorAndLower(t *testing.T) {
	if !inRange(2, 2, 6) || !inRange(4, 2, 6) || inRange(1, 2, 6) || inRange(7, 2, 6) {
		t.Fatal("interior/lower/outside behavior changed")
	}
}

func TestUpperBoundIncluded(t *testing.T) {
	if !inRange(6, 2, 6) {
		t.Fatal("upper bound should be included")
	}
}
