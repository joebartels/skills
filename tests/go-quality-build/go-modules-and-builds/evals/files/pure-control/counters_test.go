package counters

import "testing"

func TestPositive(t *testing.T) {
	if got := SumPositive([]int{2, 3}); got != 5 {
		t.Fatalf("got %d; want 5", got)
	}
}
