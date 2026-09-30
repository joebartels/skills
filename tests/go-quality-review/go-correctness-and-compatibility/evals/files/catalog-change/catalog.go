package catalog

import "sync"

type Catalog struct {
	mu  sync.RWMutex
	ids []string
}

func New(ids []string) *Catalog {
	return &Catalog{ids: append([]string(nil), ids...)}
}

// IDs returns IDs in stable order as a caller-owned snapshot.
// Callers may modify the result without affecting the Catalog.
func (c *Catalog) IDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ids[:len(c.ids):len(c.ids)]
}
