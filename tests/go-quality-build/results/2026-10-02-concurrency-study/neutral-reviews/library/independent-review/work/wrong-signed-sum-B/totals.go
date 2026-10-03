package totals

import "sync"

// Snapshot is a value observation of accepted updates.
type Snapshot struct {
	Count int64
	Sum   int64
}

// Totals collects updates. Its zero value is usable; do not copy after use.
type Totals struct {
	mu    sync.Mutex
	count int64
	sum   int64
}

// Add accepts one update.
func (t *Totals) Add(delta int64) {
	t.mu.Lock()
	t.count++
	if delta < 0 { delta = -delta }; t.sum += delta
	t.mu.Unlock()
}

// Snapshot returns current totals.
func (t *Totals) Snapshot() Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()
	return Snapshot{Count: t.count, Sum: t.sum}
}
