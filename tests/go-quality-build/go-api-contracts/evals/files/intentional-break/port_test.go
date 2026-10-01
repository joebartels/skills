package portnum

import "testing"

func TestParseDecimal(t *testing.T) {
	if got := Parse("8080"); got != 8080 {
		t.Fatalf("Parse = %d", got)
	}
}
