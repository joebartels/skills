// Package collector collects strings subject to an optional prefix and size limit.
package collector

import (
	"strings"
	"sync"
)

// Options configures a Collector created by New.
type Options struct {
	// Limit is the maximum number of stored strings. Zero means no limit.
	// Limit must be nonnegative.
	Limit int
	// Prefix is the case-sensitive prefix each stored string must have.
	// An empty prefix accepts every string.
	Prefix string
}

// Collector stores strings in the order they are successfully added.
// Its zero value accepts all strings without a limit.
// Its methods are safe for concurrent use. A Collector must not be copied
// after first use.
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

// Add stores value and returns true if it matches the configured prefix and
// the collector has room. Otherwise, it returns false without changing the
// stored strings. Duplicate strings are retained.
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

// Values returns a snapshot of the stored strings in insertion order.
// Modifying the returned slice does not change the collector's stored strings.
func (c *Collector) Values() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Keep the collector's backing array private so callers cannot mutate it.
	values := make([]string, len(c.values))
	copy(values, c.values)
	return values
}
