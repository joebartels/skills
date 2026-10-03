// Package collector collects strings that match a prefix, with an optional limit.
package collector

import (
	"strings"
	"sync"
)

// Options configures a Collector created by New.
type Options struct {
	// Limit is the maximum number of strings to collect. Zero means no limit.
	// A negative Limit causes New to panic.
	Limit int
	// Prefix is the required prefix for accepted strings. An empty Prefix
	// accepts all strings.
	Prefix string
}

// Collector stores accepted strings, including duplicates, in insertion order.
// Its methods are safe for concurrent use. The zero value accepts all strings
// without a limit. A Collector must not be copied after first use.
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

// Add appends value and reports whether it was accepted. It returns false
// without changing the collection if value does not match the configured prefix
// or the collection has reached its positive limit.
func (c *Collector) Add(value string) bool {
	// Keep the limit check and append under the same lock so concurrent calls
	// cannot exceed the limit.
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

// Values returns a snapshot of the collected strings in insertion order.
// The caller may modify the returned slice without affecting the Collector.
func (c *Collector) Values() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Copy while holding the lock to avoid exposing the collector's backing slice.
	values := make([]string, len(c.values))
	copy(values, c.values)
	return values
}
