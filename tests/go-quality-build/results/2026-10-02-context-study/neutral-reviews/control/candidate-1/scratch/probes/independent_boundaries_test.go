package rangecheck

import "testing"

func TestIndependentBoundaries(t *testing.T) {
    max := int(^uint(0) >> 1)
    min := -max - 1
    cases := []struct {
        name string
        value, low, high int
        want bool
    }{
        {"singleton included", 3, 3, 3, true},
        {"singleton below", 2, 3, 3, false},
        {"singleton above", 4, 3, 3, false},
        {"negative lower", -6, -6, -2, true},
        {"negative upper", -2, -6, -2, true},
        {"negative interior", -4, -6, -2, true},
        {"negative below", -7, -6, -2, false},
        {"negative above", -1, -6, -2, false},
        {"zero singleton", 0, 0, 0, true},
        {"minimum singleton", min, min, min, true},
        {"maximum singleton", max, max, max, true},
        {"full range minimum", min, min, max, true},
        {"full range maximum", max, min, max, true},
        {"full range zero", 0, min, max, true},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            if got := inRange(c.value, c.low, c.high); got != c.want {
                t.Fatalf("inRange(%d, %d, %d) = %v; want %v", c.value, c.low, c.high, got, c.want)
            }
        })
    }
}
