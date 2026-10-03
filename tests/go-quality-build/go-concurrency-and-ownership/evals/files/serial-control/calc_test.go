package serialcalc

import "testing"

func TestZero(t *testing.T) {
	if got := SquareSum(nil); got != 0 {
		t.Fatalf("got=%d", got)
	}
}
