// Package keytext defines catalog name formatting policy.
package keytext

import "strings"

// Display removes surrounding whitespace from a displayed name.
func Display(name string) string {
	return strings.TrimSpace(name)
}
