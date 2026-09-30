package catalog

// Catalog owns its labels. It is used sequentially by callers.
type Catalog struct{ labels []string }

func New(labels []string) *Catalog { return &Catalog{labels: append([]string(nil), labels...)} }

// Snapshot returns independently mutable labels.
func (c *Catalog) Snapshot() []string { return append([]string(nil), c.labels...) }
