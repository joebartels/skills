package rangecheck

import "testing"

func TestInclusiveUpperBound(t *testing.T) {
	for _, c := range []struct {
		value int
		want  bool
	}{{1, false}, {2, true}, {4, true}, {6, true}, {7, false}} {
		if got := inRange(c.value, 2, 6); got != c.want {
			t.Fatalf("value=%d got=%v want=%v", c.value, got, c.want)
		}
	}
}
