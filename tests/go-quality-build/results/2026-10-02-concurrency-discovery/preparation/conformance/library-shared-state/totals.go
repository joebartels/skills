package totals

import "sync"

type Snapshot struct {
	Count int64
	Sum   int64
}
type Totals struct {
	mu         sync.Mutex
	count, sum int64
}

func (t *Totals) Add(delta int64) { t.mu.Lock(); defer t.mu.Unlock(); t.count++; t.sum += delta }
func (t *Totals) Snapshot() Snapshot {
	t.mu.Lock()
	defer t.mu.Unlock()
	return Snapshot{t.count, t.sum}
}
