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

// Record joins escaped fields and appends one newline.
func Record(cells []string) string {
	quoted := make([]string, len(cells))
	for i, cell := range cells {
		quoted[i] = Quote(cell)
	}
	return strings.Join(quoted, ",") + "\n"
}
