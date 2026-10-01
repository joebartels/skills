package batches

import "testing"

func TestWholeBatches(t *testing.T) {
	for _, tc := range []struct{ items, size, want int }{{0, 4, 0}, {8, 4, 2}, {3, 1, 3}} {
		if got := batchCount(tc.items, tc.size); got != tc.want {
			t.Errorf("batchCount(%d, %d) = %d; want %d", tc.items, tc.size, got, tc.want)
		}
	}
}
