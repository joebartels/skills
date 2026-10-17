package catalog_test

import (
	"testing"

	"example.org/catalog"
)

func TestLabel(t *testing.T) {
	if got := catalog.Label("  North Pier 7 "); got != "North Pier 7" {
		t.Fatalf("Label = %q, want %q", got, "North Pier 7")
	}
}
