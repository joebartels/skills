package calculation

import "testing"

func TestPositiveInput(t *testing.T) {
	if got := SumPositive([]int{2, 3}); got != 5 {
		t.Fatalf("sum = %d, want 5", got)
	}
}
