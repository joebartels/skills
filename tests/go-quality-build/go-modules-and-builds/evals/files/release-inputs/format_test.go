package codes

import "testing"

func TestValid(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		want        bool
	}{
		{"invoice", "INV-004219", true},
		{"receipt", "RCT-00000001", true},
		{"invoice", "INV-00421", false},
		{"invoice", "INV-00421x", false},
		{"invoice", "INV-１２３４５６", false},
		{"invoice", "RCT-004219", false},
		{"Invoice", "INV-004219", false},
		{"unknown", "", false},
	} {
		if got := Valid(tc.name, tc.value); got != tc.want {
			t.Errorf("Valid(%q, %q) = %v; want %v", tc.name, tc.value, got, tc.want)
		}
	}
}
