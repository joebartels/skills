package catalog
import "testing"
func TestReviewConstructorIsolation(t *testing.T) {
 input := []string{"alpha"}; c := New(input); input[0] = "changed"
 if got := c.Snapshot()[0]; got != "alpha" {t.Fatalf("constructor alias: %q", got)}
}
func TestReviewEmptyAndZero(t *testing.T) {
 for _, c := range []*Catalog{New(nil), New([]string{}), new(Catalog)} {
  if len(c.Snapshot()) != 0 { t.Fatal("nonempty snapshot") }
 }
}
func TestReviewSnapshotIndependent(t *testing.T) {
 c := New([]string{"alpha"}); first := c.Snapshot(); second := c.Snapshot(); first[0] = "changed"
 if got := c.Snapshot()[0]; got != "alpha" { t.Errorf("stored label mutated: got %q want alpha", got) }
 if second[0] != "alpha" { t.Errorf("second snapshot mutated: got %q want alpha",second[0]) }
}
