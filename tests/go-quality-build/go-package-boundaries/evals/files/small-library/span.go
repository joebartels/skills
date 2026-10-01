// Package span provides half-open integer intervals.
package span

import "fmt"

// Range contains Start and excludes End. Equal bounds denote an empty range.
type Range struct{ Start, End int }

// New rejects reversed bounds.
func New(start, end int) (Range, error) {
	if start > end {
		return Range{}, fmt.Errorf("reversed bounds")
	}
	return Range{start, end}, nil
}

// Contains reports whether n is in r.
func (r Range) Contains(n int) bool { return n >= r.Start && n < r.End }
