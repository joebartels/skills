package positive

import "testing"

func TestControllerFinalPositiveValue(t *testing.T) {
	for _, c := range []struct {
		values []int
		want   int
	}{{nil, 0}, {[]int{-3, -2}, 0}, {[]int{7}, 7}, {[]int{3, -2, 7}, 10}, {[]int{0, 2, 4}, 6}} {
		if got := sumPositive(c.values); got != c.want {
			t.Fatalf("values=%v got=%d want=%d", c.values, got, c.want)
		}
	}
}
