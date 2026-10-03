package rangecheck

import "testing"

func TestReviewerFiniteInclusiveIntervals(t *testing.T) {
    for low := -3; low <= 3; low++ {
        for high := low; high <= 3; high++ {
            for value := -4; value <= 4; value++ {
                want := false
                for member := low; member <= high; member++ {
                    if value == member { want = true }
                }
                if got := inRange(value, low, high); got != want {
                    t.Fatalf("value=%d low=%d high=%d got=%v want=%v", value, low, high, got, want)
                }
            }
        }
    }
}

func TestReviewerIntExtremes(t *testing.T) {
    max := int(^uint(0) >> 1)
    min := -max - 1
    cases := []struct { value, low, high int; want bool }{
        {min, min, min, true}, {max, max, max, true},
        {min, min, max, true}, {max, min, max, true},
        {min, min + 1, max, false}, {max, min, max - 1, false},
    }
    for _, c := range cases {
        if got := inRange(c.value, c.low, c.high); got != c.want {
            t.Fatalf("value=%d low=%d high=%d got=%v want=%v", c.value, c.low, c.high, got, c.want)
        }
    }
}
