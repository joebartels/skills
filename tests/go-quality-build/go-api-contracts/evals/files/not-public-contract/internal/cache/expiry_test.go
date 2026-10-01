package cache

import "testing"

func TestExpired(t *testing.T) {
	if !expired(12, 10) {
		t.Fatal("past deadline is live")
	}
	if expired(8, 10) {
		t.Fatal("future deadline is expired")
	}
	if expired(12, 0) {
		t.Fatal("permanent entry is expired")
	}
}
