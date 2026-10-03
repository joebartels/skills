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
	// Prefix is the required prefix of each value. An empty prefix accepts
	// any string.
	Prefix string
}

// Collector stores accepted strings in insertion order. Its methods are safe
// for concurrent use. A Collector must not be copied after first use.
// The zero value accepts all strings with no limit.
type Collector struct {
	// mu protects values and keeps the limit check atomic with appending.
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
// the limit has not been reached. Otherwise it returns false without changing
// the collected values. Duplicate values are accepted and count toward the limit.
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

// Values returns a snapshot of the collected values in insertion order.
// The caller may modify the returned slice without affecting the Collector.
func (c *Collector) Values() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	values := make([]string, len(c.values))
	copy(values, c.values)
	return values
}
