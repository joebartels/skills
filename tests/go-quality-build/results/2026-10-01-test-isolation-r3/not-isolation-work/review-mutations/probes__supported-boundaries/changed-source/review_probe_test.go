package clamp_test

import (
    "testing"
    "example.com/clamp"
)

var compatibleFunction func(int, int, int) int = clamp.Clamp

func TestReviewSupportedBoundaries(t *testing.T) {
    max := int(^uint(0) >> 1)
    min := -max - 1
    cases := []struct{ value, low, high, want int }{
        {min, -7, 3, -7}, {-7, -7, 3, -7}, {-1, -7, 3, -1},
        {3, -7, 3, 3}, {max, -7, 3, 3},
        {min, min, max, min}, {max, min, max, max},
        {-1, 0, 0, 0}, {0, 0, 0, 0}, {1, 0, 0, 0},
    }
    for _, c := range cases {
        if got := compatibleFunction(c.value, c.low, c.high); got != c.want {
            t.Errorf("Clamp(%d,%d,%d) = %d; want %d", c.value, c.low, c.high, got, c.want)
        }
    }
}
