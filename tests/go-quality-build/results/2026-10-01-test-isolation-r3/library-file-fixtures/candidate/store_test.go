package filestore_test

import (
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
		values []string
		root   string
		want   string
		wrote  bool
	}{
		{name: "alpha", values: []string{"alpha initial value", "A"}},
		{name: "beta", values: []string{"beta initial value", "B\n"}},
		{name: "gamma", values: []string{"gamma initial value", "\u03b3\x00"}},
	}
	for i := range cases {
		cases[i].root = t.TempDir()
	}

	// The group joins its parallel children before the persistence checks.
	t.Run("instances", func(t *testing.T) {
		for i := range cases {
			instance := &cases[i]
			t.Run(instance.name, func(t *testing.T) {
				t.Parallel()
				store := filestore.New(instance.root)
				for i, value := range instance.values {
					name := []string{"first", "replacement"}[i]
					t.Run(name, func(t *testing.T) {
						if err := store.Put("same-key", value); err != nil {
							t.Fatal(err)
						}
						instance.want = value
						instance.wrote = true
						got, err := store.Get("same-key")
						if err != nil || got != value {
							t.Fatalf("Get = %q, %v; want %q", got, err, value)
						}
					})
				}
			})
		}
	})

	for _, instance := range cases {
		// A focused -run may execute only one instance or one value case.
		if !instance.wrote {
			continue
		}
		got, err := filestore.New(instance.root).Get("same-key")
		if err != nil || got != instance.want {
			t.Errorf("%s persisted Get = %q, %v; want %q", instance.name, got, err, instance.want)
		}
	}
}
