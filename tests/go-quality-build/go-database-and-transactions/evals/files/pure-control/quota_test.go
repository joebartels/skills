package quotas

import "testing"

func TestPositive(t *testing.T) {
	if got := Total([]int{2, 3}); got != 5 {
		t.Fatalf("got %d", got)
	}
}
