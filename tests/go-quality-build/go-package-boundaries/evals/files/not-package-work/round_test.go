package mathutil

import "testing"

func TestExactBuckets(t *testing.T) {
	if buckets(12, 4) != 3 || buckets(0, 4) != 0 {
		t.Fatal("wrong count")
	}
}
