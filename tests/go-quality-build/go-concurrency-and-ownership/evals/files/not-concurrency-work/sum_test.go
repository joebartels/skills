package positive

import "testing"

func TestSumWithNonpositiveTail(t *testing.T) {
	if got := sumPositive([]int{2, -1, 0}); got != 2 {
		t.Fatalf("sum=%d", got)
	}
}
