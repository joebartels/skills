package positive

import "testing"

func TestReviewContractCases(t *testing.T) {
    cases := []struct {
        name string
        values []int
        want int
    }{
        {"nil", nil, 0},
        {"empty", []int{}, 0},
        {"positive singleton", []int{7}, 7},
        {"zero singleton", []int{0}, 0},
        {"negative singleton", []int{-7}, 0},
        {"negative only", []int{-9, -2, -1}, 0},
        {"zeros only", []int{0, 0, 0}, 0},
        {"positive first", []int{7, 0, -2}, 7},
        {"positive middle", []int{0, 7, -2}, 7},
        {"positive final", []int{0, -2, 7}, 7},
        {"all positive", []int{1, 2, 3, 4}, 10},
        {"mixed", []int{-5, 4, 0, 2, -1, 3}, 9},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            before := append([]int(nil), tc.values...)
            if got := sumPositive(tc.values); got != tc.want {
                t.Fatalf("sumPositive(%v)=%d, want %d", tc.values, got, tc.want)
            }
            for i, value := range before {
                if tc.values[i] != value {
                    t.Fatalf("input changed at index %d", i)
                }
            }
        })
    }
}
