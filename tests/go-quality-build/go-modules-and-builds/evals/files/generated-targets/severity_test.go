package severity

import "testing"

func TestExistingNames(t *testing.T) {
	for _, tc := range []struct {
		code int
		name string
	}{{10, "debug"}, {20, "info"}, {-1, ""}} {
		if got := Name(tc.code); got != tc.name {
			t.Errorf("Name(%d) = %q; want %q", tc.code, got, tc.name)
		}
	}
}
