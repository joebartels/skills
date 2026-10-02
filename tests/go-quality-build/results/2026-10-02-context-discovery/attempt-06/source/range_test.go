package rangecheck

import "testing"

func TestInclusiveInterval(t *testing.T) {
	if !inRange(2, 2, 6) || !inRange(4, 2, 6) || !inRange(6, 2, 6) || inRange(1, 2, 6) || inRange(7, 2, 6) {
		t.Fatal("inclusive interval behavior changed")
	}
}
