// Package collector stores strings that match a configured prefix, optionally
// limiting the number of stored values.
package collector

import (
	"strings"
	"sync"
)

// Options configures a Collector created by New.
type Options struct {
	// Limit is the maximum number of stored values. Zero means no limit.
	// New panics if Limit is negative.
	Limit int
	// Prefix is the case-sensitive prefix required for a value to be accepted.
	// An empty prefix accepts all strings.
	Prefix string
}

// Collector stores accepted strings in insertion order, including duplicates.
// Its zero value is ready to use and accepts all strings without a limit.
// Its methods are safe for concurrent use. A Collector must not be copied after
// first use.
type Collector struct {
	mu     sync.Mutex // Protects values; limit and prefix do not change after creation.
	limit  int
	prefix string
	values []string
}

// New returns a Collector configured with options.
// It panics if options.Limit is negative.
func New(options Options) *Collector {
	if options.Limit < 0 {
		panic("negative limit")
	}
	return &Collector{limit: options.Limit, prefix: options.Prefix}
}

// Add stores value and reports whether it was accepted. It returns false without
// storing value if the prefix does not match or the configured limit is reached.
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

// Values returns a snapshot of the stored values in insertion order.
// Modifying the returned slice does not affect the Collector.
func (c *Collector) Values() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	values := make([]string, len(c.values))
	copy(values, c.values)
	return values
}
