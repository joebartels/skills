package paging

import "testing"

func TestPartial(t *testing.T) {
	if pages(3, 2) != 2 {
		t.Fatal("partial page")
	}
}
