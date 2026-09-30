package catalog

import "testing"

func TestIDsCanBeAppendedTo(t *testing.T) {
	c := New([]string{"alpha", "beta"})
	got := append(c.IDs(), "gamma")
	if len(got) != 3 || len(c.IDs()) != 2 {
		t.Fatalf("unexpected IDs: result=%v catalog=%v", got, c.IDs())
	}
}
