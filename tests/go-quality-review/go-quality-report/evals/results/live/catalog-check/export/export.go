package export

import (
	"example.com/catalog-export/catalog"
	"strings"
)

// Uppercase exports labels without modifying their catalog.
func Uppercase(c *catalog.Catalog) string {
	labels := c.Snapshot()
	for i := range labels {
		labels[i] = strings.ToUpper(labels[i])
	}
	return strings.Join(labels, ",")
}
