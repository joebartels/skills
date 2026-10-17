package tally_test

import (
	"testing"

	tally "example.com/oddtally"
)

func TestTotalPositiveOdds(t *testing.T) {
	if got := tally.Total([]int{0, 1, 2, 3, 4}); got != 4 {
		t.Fatalf("Total = %d, want 4", got)
	}
}
