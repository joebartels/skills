// Package collector collects strings that match a prefix, with an optional limit.
package collector

import (
	"strings"
	"sync"
)

// Options configures a Collector created by New.
type Options struct {
	// Limit is the maximum number of values to collect. Zero means no limit.
	// A negative limit causes New to panic.
	Limit int
	// Prefix is the case-sensitive prefix required for each value.
	// An empty prefix matches every string.
	Prefix string
}

// Collector stores accepted strings in the order they are added, including
// duplicates. Its methods are safe for concurrent use. The zero value is ready
// to use and accepts all strings without a limit.
//
// A Collector must not be copied after first use.
type Collector struct {
	mu     sync.Mutex
	limit  int
	prefix string
	values []string
}

// New returns an empty Collector configured with options.
// It panics if options.Limit is negative.
func New(options Options) *Collector {
	if options.Limit < 0 {
		panic("negative limit")
	}
	return &Collector{limit: options.Limit, prefix: options.Prefix}
}

// Add appends value if it matches the configured prefix and the collector has
// not reached its limit. It reports whether value was added. Rejected values
// leave the collector unchanged.
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

// Values returns a snapshot of the collected strings in the order they were
// added. The caller may modify the returned slice without affecting the
// collector, and later calls to Add do not change the returned slice.
func (c *Collector) Values() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Keep the collector's backing array private so callers cannot mutate it.
	values := make([]string, len(c.values))
	copy(values, c.values)
	return values
}
