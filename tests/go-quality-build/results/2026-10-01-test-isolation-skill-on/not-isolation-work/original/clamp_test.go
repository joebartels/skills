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
