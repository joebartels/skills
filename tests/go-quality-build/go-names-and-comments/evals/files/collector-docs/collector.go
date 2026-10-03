package collector

import (
	"strings"
	"sync"
)

type Options struct {
	Limit  int
	Prefix string
}

type Collector struct {
	mu     sync.Mutex
	limit  int
	prefix string
	values []string
}

func New(options Options) *Collector {
	if options.Limit < 0 {
		panic("negative limit")
	}
	return &Collector{limit: options.Limit, prefix: options.Prefix}
}

func (c *Collector) Add(value string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !strings.HasPrefix(value, c.prefix) {
		return false
	}
	if c.limit > 0 && len(c.values) >= c.limit {
		return false
	}
	c.values = append(c.values, value)
	return true
}

func (c *Collector) Values() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	values := make([]string, len(c.values))
	copy(values, c.values)
	return values
}
