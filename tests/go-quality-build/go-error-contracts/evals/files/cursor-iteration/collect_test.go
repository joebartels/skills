package cursorload

import "testing"

type oneRow struct{ done bool }

func (r *oneRow) Next() bool {
	if r.done {
		return false
	}
	r.done = true
	return true
}
func (*oneRow) Scan(s *string) error { *s = "a"; return nil }
func (*oneRow) Err() error           { return nil }
func (*oneRow) Close() error         { return nil }
func TestCollect(t *testing.T) {
	got, err := Collect(new(oneRow))
	if err != nil || len(got) != 1 || got[0] != "a" {
		t.Fatal(got, err)
	}
}
