package export
import (
 "testing"
 "example.com/catalog-export/catalog"
)
func TestReviewUppercasePreservesCatalog(t *testing.T) {
 c := catalog.New([]string{"alpha"})
 Uppercase(c)
 if got := c.Snapshot()[0]; got != "alpha" { t.Fatalf("catalog mutated: got %q, want alpha", got) }
}
func TestReviewSnapshotElementIsolation(t *testing.T) {
 c := catalog.New([]string{"alpha"})
 c.Snapshot()[0] = "changed"
 if got := c.Snapshot()[0]; got != "alpha" { t.Fatalf("snapshot write mutated catalog: %q", got) }
}
