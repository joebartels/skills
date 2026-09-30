package limits

import "testing"

func TestValidLimit(t *testing.T) {
	if err := ValidateLimit(10); err != nil {
		t.Fatalf("valid input returned non-nil error: %v", err)
	}
}
