package client

import "testing"

func TestOnlyPositiveContributes(t *testing.T) {
	if got := SumPositive([]int{-9, 0, 4, 2}); got != 6 {
		t.Fatalf("got %d want 6", got)
	}
}
