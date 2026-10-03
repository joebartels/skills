// Package collector collects strings that match a prefix, with an optional limit.
package collector

import (
	"strings"
	"sync"
)

// Options configures a Collector created by New.
type Options struct {
	// Limit is the maximum number of strings to store. Zero means unlimited.
	// A negative limit causes New to panic.
	Limit int
	// Prefix is the required string prefix. An empty prefix matches every string.
	Prefix string
}

// Collector stores strings in the order they are added.
// The zero value is ready for use, accepts every string, and has no limit.
// A Collector is safe for concurrent use and must not be copied after first use.
type Collector struct {
	// mu protects values; limit and prefix do not change after construction.
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

// Add appends value and returns true if it matches the configured prefix and
// the collector has room. Otherwise, it returns false without changing the collector.
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

// Values returns a copy of the stored strings in the order they were added.
// Modifying the returned slice does not affect the collector.
func (c *Collector) Values() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	values := make([]string, len(c.values))
	copy(values, c.values)
	return values
}
