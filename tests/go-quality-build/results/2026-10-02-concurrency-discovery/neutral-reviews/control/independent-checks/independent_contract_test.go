package positive

import (
    "reflect"
    "testing"
)

func TestIndependentContractBoundaries(t *testing.T) {
    cases := []struct {
        name string
        values []int
        want int
    }{
        {"nil", nil, 0},
        {"empty", []int{}, 0},
        {"negative-only", []int{-3, -2, -1}, 0},
        {"zero-only", []int{0, 0}, 0},
        {"single-positive", []int{7}, 7},
        {"single-negative", []int{-7}, 0},
        {"single-zero", []int{0}, 0},
        {"only-final-positive", []int{-2, 0, 9}, 9},
        {"all-positive", []int{1, 2, 3, 4}, 10},
        {"mixed", []int{4, -9, 0, 7, -1, 3}, 14},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            before := make([]int, len(tc.values))
            copy(before, tc.values)
            if tc.values == nil { before = nil }
            if got := sumPositive(tc.values); got != tc.want {
                t.Fatalf("sumPositive(%v)=%d, want %d", tc.values, got, tc.want)
            }
            if !reflect.DeepEqual(tc.values, before) { t.Fatalf("input mutated: got %v, before %v", tc.values, before) }
        })
    }
}

func TestIndependentBoundedEnumeration(t *testing.T) {
    count := 0
    var visit func([]int, int)
    visit = func(values []int, want int) {
        count++
        if got := sumPositive(values); got != want { t.Fatalf("sumPositive(%v)=%d, want %d", values, got, want) }
        if len(values) == 5 { return }
        for x := -2; x <= 2; x++ {
            next := append(append([]int(nil), values...), x)
            nextWant := want
            if x > 0 { nextWant += x }
            visit(next, nextWant)
        }
    }
    visit(nil, 0)
    if count != 3906 { t.Fatalf("review enumeration count=%d, want 3906", count) }
    t.Logf("checked %d bounded input vectors", count)
}
