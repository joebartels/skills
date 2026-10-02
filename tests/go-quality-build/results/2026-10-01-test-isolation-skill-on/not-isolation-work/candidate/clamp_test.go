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
	for _, tc := range []struct {
		name  string
		value int
	}{
		{name: "below", value: 0},
		{name: "inclusive", value: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := clamp.Clamp(tc.value, 1, 9); got != 1 {
				t.Fatalf("Clamp(%d, 1, 9) = %d, want 1", tc.value, got)
			}
		})
	}
}
