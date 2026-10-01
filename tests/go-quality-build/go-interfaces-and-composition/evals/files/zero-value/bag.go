package bag

type Bag struct{ counts map[string]int }

func (b *Bag) Add(key string) {
	if b.counts == nil {
		b.counts = make(map[string]int)
	}
	b.counts[key]++
}
func (b *Bag) Count(key string) int { return b.counts[key] }
