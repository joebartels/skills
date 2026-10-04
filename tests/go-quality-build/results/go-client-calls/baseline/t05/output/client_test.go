package client

import "testing"

func TestExistingPositives(t *testing.T) {
	if got := SumPositive([]int{1, 2, 3}); got != 6 {
		t.Fatalf("got %d", got)
	}
}

func TestSumPositiveIgnoresNonPositive(t *testing.T) {
	if got := SumPositive([]int{-5, 2, 0, -1, 3}); got != 5 {
		t.Fatalf("SumPositive() = %d, want 5", got)
	}
}
