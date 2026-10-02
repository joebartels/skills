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
	tests := []struct {
		name  string
		value int
		low   int
		high  int
	}{
		{name: "below", value: 0, low: 1, high: 9},
		{name: "at", value: 1, low: 1, high: 9},
		{name: "single value interval", value: 1, low: 1, high: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clamp.Clamp(tt.value, tt.low, tt.high); got != tt.low {
				t.Fatalf("Clamp(%d, %d, %d) = %d, want %d", tt.value, tt.low, tt.high, got, tt.low)
			}
		})
	}
}
