package totals

// Snapshot is a value observation of accepted updates.
type Snapshot struct {
	Count int64
	Sum   int64
}

// Totals collects updates. Its zero value is usable; do not copy after use.
type Totals struct {
	count int64
	sum   int64
}

// Add accepts one update.
func (t *Totals) Add(delta int64) { t.count++; t.sum += delta }

// Snapshot returns current totals.
func (t *Totals) Snapshot() Snapshot { return Snapshot{Count: t.count, Sum: t.sum} }
