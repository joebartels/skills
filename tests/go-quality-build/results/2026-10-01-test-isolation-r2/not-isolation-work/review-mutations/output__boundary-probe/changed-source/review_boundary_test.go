package clamp_test

import (
    "testing"
    "example.com/clamp"
)

func TestReviewSupportedBoundaries(t *testing.T) {
    max := int(^uint(0) >> 1)
    min := -max - 1
    cases := []struct{ value, low, high, want int }{
        {-8, -5, 3, -5}, {-5, -5, 3, -5}, {3, -5, 3, 3},
        {8, -5, 3, 3}, {7, 7, 7, 7}, {6, 7, 7, 7}, {8, 7, 7, 7},
        {min, min, max, min}, {max, min, max, max},
        {min, min+1, max, min+1}, {max, min, max-1, max-1},
    }
    for _, tc := range cases {
        if got := clamp.Clamp(tc.value, tc.low, tc.high); got != tc.want {
            t.Errorf("Clamp(%d,%d,%d)=%d,want %d", tc.value,tc.low,tc.high,got,tc.want)
        }
    }
}
