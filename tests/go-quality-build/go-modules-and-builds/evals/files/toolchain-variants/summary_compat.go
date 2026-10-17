//go:build !go1.23

package roster

import "sort"

// Summarize reports the total and sorted distinct nonempty names.
func Summarize(entries []Entry) Summary {
	names := make(map[string]struct{})
	for _, entry := range entries {
		if entry.Name != "" {
			names[entry.Name] = struct{}{}
		}
	}
	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	return Summary{Total: len(entries), Names: sorted}
}
