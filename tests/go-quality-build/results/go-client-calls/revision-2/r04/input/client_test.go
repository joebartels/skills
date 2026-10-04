package client

import "testing"

func TestExistingPositives(t *testing.T) {
	if got := SumPositive([]int{1, 2, 3}); got != 6 {
		t.Fatalf("got %d", got)
	}
}
