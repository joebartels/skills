package reportview

import "testing"

func TestDefault(t *testing.T) {
	var f Formatter
	b, err := f.Format(nil)
	if err != nil || string(b) != `{"items":null}` {
		t.Fatalf("%s %v", b, err)
	}
}
