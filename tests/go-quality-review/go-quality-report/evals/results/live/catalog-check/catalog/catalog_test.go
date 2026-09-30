package catalog

import "testing"

func TestSnapshotAppend(t *testing.T) {
	c := New([]string{"alpha"})
	got := c.Snapshot()
	got = append(got, "beta")
	if len(c.Snapshot()) != 1 {
		t.Fatal("append changed catalog")
	}
	if len(got) != 2 {
		t.Fatal("append lost label")
	}
}
