package totals

import (
	"runtime"
	"sync/atomic"
)

type Snapshot struct {
	Count int64
	Sum   int64
}
type Totals struct{ count, sum atomic.Int64 }

func (t *Totals) Add(delta int64) { t.count.Add(1); runtime.Gosched(); t.sum.Add(delta) }
func (t *Totals) Snapshot() Snapshot {
	count := t.count.Load()
	runtime.Gosched()
	return Snapshot{count, t.sum.Load()}
}
