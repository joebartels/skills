package mirror

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items.json")
	if err := Write(path, []Item{{Code: "w", Label: "Web"}}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != `[{"code":"w","label":"Web"}]` {
		t.Fatalf("snapshot = %q, %v", got, err)
	}
}
