package client

import "testing"

func TestExistingPositives(t *testing.T) {
	if got := SumPositive([]int{1, 2, 3}); got != 6 {
		t.Fatalf("got %d", got)
	}
}

func TestSumPositiveIgnoresNonPositiveValues(t *testing.T) {
	var sumPositive func([]int) int = SumPositive
	if got := sumPositive([]int{-3, 1, 0, 2, -4, 3}); got != 6 {
		t.Fatalf("SumPositive() = %d, want 6", got)
	}
}
