package bag

import "testing"

func TestZeroBag(t *testing.T) {
	var b Bag
	if got := b.Count("missing"); got != 0 {
		t.Fatalf("missing = %d", got)
	}
	b.Add("jobs")
	b.Add("jobs")
	b.Add("")
	if got := b.Count("jobs"); got != 2 {
		t.Fatalf("jobs = %d", got)
	}
	if got := b.Count(""); got != 1 {
		t.Fatalf("empty key = %d", got)
	}
	var other Bag
	if got := other.Count("jobs"); got != 0 {
		t.Fatalf("other bag = %d", got)
	}
}
