//go:build go1.23

package roster

import (
	"maps"
	"slices"
)

// Summarize reports the total and sorted distinct nonempty names.
func Summarize(entries []Entry) Summary {
	names := make(map[string]struct{})
	for _, entry := range entries {
		if entry.Name != "" {
			names[entry.Name] = struct{}{}
		}
	}
	sorted := slices.Sorted(maps.Keys(names))
	if sorted == nil {
		sorted = []string{}
	}
	return Summary{Total: len(entries), Names: sorted}
}
