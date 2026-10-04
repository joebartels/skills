package paging

// pages counts pages for nonnegative items and a positive size.
func pages(items, size int) int {
	count := items / size
	if items%size == 0 {
		return count
	}
	return count + 1
}
