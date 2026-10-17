// Package catalog formats human-facing catalog names.
package catalog

import "example.org/keytext"

// Label trims surrounding whitespace while preserving the displayed name.
func Label(name string) string {
	return keytext.Display(name)
}
