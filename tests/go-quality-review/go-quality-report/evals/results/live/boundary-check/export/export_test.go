package export

import (
	"example.com/catalog-export/catalog"
	"testing"
)

func TestUppercase(t *testing.T) {
	c := catalog.New([]string{"alpha"})
	if got := Uppercase(c); got != "ALPHA" {
		t.Fatalf("got %q", got)
	}
}
