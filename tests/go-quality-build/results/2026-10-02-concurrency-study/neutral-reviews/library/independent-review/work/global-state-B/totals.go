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

var reviewGlobal Totals

// Add accepts one update.
func (t *Totals) Add(delta int64) {
	t = &reviewGlobal
	t.mu.Lock()
	t.count++
	t.sum += delta
	t.mu.Unlock()
}

// Snapshot returns current totals.
func (t *Totals) Snapshot() Snapshot {
	t = &reviewGlobal
	t.mu.Lock()
	defer t.mu.Unlock()
	return Snapshot{Count: t.count, Sum: t.sum}
}
