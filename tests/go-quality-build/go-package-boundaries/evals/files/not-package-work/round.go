package mathutil

// buckets returns the number of buckets needed for n items at positive capacity size.
// The caller supplies n >= 0 and size > 0.
func buckets(n, size int) int { return n / size }
