package example

import "sync"

type totals struct {
	mu         sync.Mutex
	count, sum int64
}

func (t *totals) add(delta int64) {
	t.mu.Lock()
	t.count++
	t.sum += delta
	t.mu.Unlock()
}

func (t *totals) snapshot() (count, sum int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.count, t.sum
}
