package client

import "testing"

func TestExistingPositives(t *testing.T) {
	if got := SumPositive([]int{1, 2, 3}); got != 6 {
		t.Fatalf("got %d", got)
	}
}

func TestSumPositiveIgnoresNonPositive(t *testing.T) {
	var sum func([]int) int = SumPositive
	if got := sum([]int{-4, 0, 2, -3, 5}); got != 7 {
		t.Fatalf("SumPositive() = %d, want 7", got)
	}
}
