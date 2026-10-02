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
	instances := []struct {
		name        string
		replacement string
		store       *filestore.Store
		want        string
		wrote       bool
	}{
		{name: "alpha", replacement: "alpha replacement"},
		{name: "beta", replacement: "beta\nreplacement\x00"},
		{name: "gamma", replacement: "gamma replacement"},
	}
	for i := range instances {
		// These roots belong to the outer test, which reads them after the group.
		instances[i].store = filestore.New(t.TempDir())
	}

	// Returning from this group waits for all selected parallel descendants.
	t.Run("instances", func(t *testing.T) {
		for i := range instances {
			instance := &instances[i]
			t.Run(instance.name, func(t *testing.T) {
				t.Parallel()
				values := []struct {
					name  string
					value string
				}{
					{name: "initial", value: "shared value"},
					{name: "repeat", value: "shared value"},
					{name: "replacement", value: instance.replacement},
				}
				// Same-root replacement remains serial inside each independent instance.
				for _, value := range values {
					t.Run(value.name, func(t *testing.T) {
						if err := instance.store.Put("same-key", value.value); err != nil {
							t.Fatalf("Put: %v", err)
						}
						instance.want = value.value
						instance.wrote = true
						got, err := instance.store.Get("same-key")
						if err != nil {
							t.Fatalf("Get: %v", err)
						}
						if got != value.value {
							t.Fatalf("Get = %q; want %q", got, value.value)
						}
					})
				}
			})
		}
	})

	for _, instance := range instances {
		// A focused -run may exclude an instance or some of its value cases.
		if !instance.wrote {
			continue
		}
		got, err := instance.store.Get("same-key")
		if err != nil {
			t.Errorf("%s: Get after children: %v", instance.name, err)
			continue
		}
		if got != instance.want {
			t.Errorf("%s: Get after children = %q; want %q", instance.name, got, instance.want)
		}
	}
}
