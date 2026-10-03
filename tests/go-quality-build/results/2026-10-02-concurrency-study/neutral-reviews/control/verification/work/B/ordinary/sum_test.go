package positive

import "testing"

func TestSumWithNonpositiveTail(t *testing.T) {
	if got := sumPositive([]int{2, -1, 0}); got != 2 {
		t.Fatalf("sum=%d", got)
	}
}

func TestSumIncludesPositiveFinalElement(t *testing.T) {
	if got := sumPositive([]int{2, -1, 3}); got != 5 {
		t.Fatalf("sum=%d, want 5", got)
	}
}
