package filestore_test

import (
	"os"
	"testing"

	"example.com/filestore"
)

func TestStoreSerialChildren(t *testing.T) {
	root, err := os.MkdirTemp("", "store-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	store := filestore.New(root)
	for _, value := range []string{"first", "replacement"} {
		t.Run(value, func(t *testing.T) {
			if err := store.Put("same-key", value); err != nil {
				t.Fatal(err)
			}
			got, err := store.Get("same-key")
			if err != nil || got != value {
				t.Fatalf("Get = %q, %v; want %q", got, err, value)
			}
		})
	}
}
