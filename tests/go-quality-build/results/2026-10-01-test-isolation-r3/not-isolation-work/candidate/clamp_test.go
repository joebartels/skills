package clamp_test

import (
	"testing"

	"example.com/clamp"
)

func TestInterior(t *testing.T) {
	if got := clamp.Clamp(5, 1, 9); got != 5 {
		t.Fatalf("Clamp = %d, want 5", got)
	}
}

func TestLowerBound(t *testing.T) {
	for _, value := range []int{0, 1} {
		if got := clamp.Clamp(value, 1, 9); got != 1 {
			t.Errorf("Clamp(%d, 1, 9) = %d, want 1", value, got)
		}
	}
}
