// Package celltext defines CSV field escaping policy.
package celltext

import "strings"

// Quote escapes a CSV field, retaining the field's exact contents.
func Quote(value string) string {
	if !strings.ContainsAny(value, ",\"\r\n") {
		return value
	}
	return "\"" + strings.ReplaceAll(value, "\"", "\"\"") + "\""
}
