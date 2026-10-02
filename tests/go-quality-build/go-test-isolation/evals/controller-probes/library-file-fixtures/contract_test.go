package filestore_test

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"example.com/filestore"
)

func TestParallelFixtureLifetime(t *testing.T) {
	parent := t.TempDir()
	var completed atomic.Int32
	t.Cleanup(func() {
		if got := completed.Load(); got != 3 {
			t.Errorf("completed children = %d, want 3", got)
		}
	})
	for _, value := range []string{"east", "west", "north"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()
			root := filepath.Join(parent, value)
			// Child-owned mutable state borrows only a live ancestor directory.
			if err := os.Mkdir(root, 0o700); err != nil {
				t.Fatal(err)
			}
			store := filestore.New(root)
			if err := store.Put("same", value); err != nil {
				t.Fatal(err)
			}
			if got, err := store.Get("same"); err != nil || got != value {
				t.Fatalf("child value = %q, %v; want %q", got, err, value)
			}
			completed.Add(1)
		})
	}
}
