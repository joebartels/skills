package label

import "testing"

func TestFormat(t *testing.T) {
	if got := Format("web"); got != "[web]" {
		t.Fatalf("Format = %q", got)
	}
}
