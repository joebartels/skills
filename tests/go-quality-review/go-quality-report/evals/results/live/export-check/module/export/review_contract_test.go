package export
import (
 "testing"
 "example.com/catalog-export/catalog"
)
func TestReviewUppercasePreservesCatalog(t *testing.T) {
 c := catalog.New([]string{"alpha", "Beta"})
 if got := Uppercase(c); got != "ALPHA,BETA" { t.Fatalf("export = %q", got) }
 if got := c.Snapshot(); got[0] != "alpha" || got[1] != "Beta" { t.Fatalf("catalog mutated: %q", got) }
}
