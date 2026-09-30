package catalog
import "testing"
func TestReviewSnapshotIndependent(t *testing.T) {
 c := New([]string{"alpha"})
 c.Snapshot()[0] = "changed"
 if got := c.Snapshot()[0]; got != "alpha" { t.Fatalf("catalog mutated: %q", got) }
}
