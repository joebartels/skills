// Package collector collects strings that match a prefix, with an optional limit.
package collector

import (
	"strings"
	"sync"
)

// Options configures a Collector created by New.
type Options struct {
	// Limit is the maximum number of stored values. Zero means no limit.
	Limit int
	// Prefix is the case-sensitive prefix required for accepted values.
	// An empty prefix accepts any string.
	Prefix string
}

// Collector stores strings accepted by Add and is safe for concurrent use.
// Its zero value is ready to use with no limit or prefix restriction.
// A Collector must not be copied after first use.
type Collector struct {
	mu     sync.Mutex
	limit  int
	prefix string
	values []string
}

// New returns an empty Collector configured by options.
// It panics if options.Limit is negative.
func New(options Options) *Collector {
	if options.Limit < 0 {
		panic("negative limit")
	}
	return &Collector{limit: options.Limit, prefix: options.Prefix}
}

// Add stores value and reports whether it was accepted.
// It returns false if value does not match the prefix or the limit is reached.
// Duplicate values are stored separately and count toward the limit.
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

// Values returns a snapshot of the stored values in the order they were added.
// The caller may modify the returned slice without affecting the Collector.
func (c *Collector) Values() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	values := make([]string, len(c.values))
	copy(values, c.values)
	return values
}
