package filestore_test

import (
	"os"
	"path/filepath"
	"testing"

	"example.com/filestore"
)

func TestStoreSerialChildren(t *testing.T) {
	store := filestore.New(t.TempDir())
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

func TestStoreIndependentInstances(t *testing.T) {
	cases := []struct {
		name   string
		root   string
		values []string
	}{
		{name: "alpha", values: []string{"alpha-initial", "alpha", "alpha"}},
		{name: "beta", values: []string{"beta-initial", "beta", "beta"}},
		{name: "gamma", values: []string{"gamma-initial", "", ""}},
	}
	for i := range cases {
		// The outer test also reads these roots after the parallel group finishes.
		cases[i].root = t.TempDir()
	}

	const key = "same-key"
	type writtenValue struct {
		name string
		root string
		want string
	}
	written := make(chan writtenValue, len(cases))
	t.Run("instances", func(t *testing.T) {
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				store := filestore.New(tc.root)
				var want string
				var didWrite bool
				for i, name := range []string{"initial", "replacement", "repeat"} {
					value := tc.values[i]
					t.Run(name, func(t *testing.T) {
						if err := store.Put(key, value); err != nil {
							t.Fatalf("Put(%q, %q): %v", key, value, err)
						}
						got, err := store.Get(key)
						if err != nil || got != value {
							t.Fatalf("Get(%q) = %q, %v; want %q", key, got, err, value)
						}
						want = value
						didWrite = true
					})
				}
				if didWrite {
					written <- writtenValue{name: tc.name, root: tc.root, want: want}
				}
			})
		}
	})
	close(written)

	// The serial group joins its parallel children. Each result describes only
	// selected child cases, so focused -run selections need no excluded siblings.
	for result := range written {
		got, err := filestore.New(result.root).Get(key)
		if err != nil || got != result.want {
			t.Errorf("%s: reopened Get(%q) = %q, %v; want %q", result.name, key, got, err, result.want)
		}
		data, err := os.ReadFile(filepath.Join(result.root, key))
		if err != nil || string(data) != result.want {
			t.Errorf("%s: file = %q, %v; want %q", result.name, data, err, result.want)
		}
	}
}
