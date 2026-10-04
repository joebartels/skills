package paging

import "testing"

func TestPages(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	tests := []struct {
		name  string
		items int
		size  int
		want  int
	}{
		{name: "zero items", items: 0, size: 2, want: 0},
		{name: "smaller than a page", items: 1, size: 2, want: 1},
		{name: "exact page", items: 2, size: 2, want: 1},
		{name: "exact multiple", items: 4, size: 2, want: 2},
		{name: "partial page", items: 3, size: 2, want: 2},
		{name: "one item per page", items: 3, size: 1, want: 3},
		{name: "largest exact page", items: maxInt, size: maxInt, want: 1},
		{name: "largest partial page", items: maxInt, size: maxInt - 1, want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pages(tt.items, tt.size); got != tt.want {
				t.Errorf("pages(%d, %d) = %d; want %d", tt.items, tt.size, got, tt.want)
			}
		})
	}
}
