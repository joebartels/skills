package recordfmt_test

import (
	"example.com/recordfmt"
	"testing"
)

func TestFormat(t *testing.T) {
	f := recordfmt.New("report/")
	if got := f.Format("status", "ready"); got != "report/status:ready" {
		t.Fatalf("Format = %q", got)
	}
}
