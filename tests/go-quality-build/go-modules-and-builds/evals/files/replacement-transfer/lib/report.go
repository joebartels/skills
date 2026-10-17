// Package reportcsv formats report cells for CSV output.
package reportcsv

import "example.org/celltext"

// Cell returns a CSV field, quoting it when required.
func Cell(value string) string {
	return celltext.Quote(value)
}
