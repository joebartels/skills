package reportcsv_test

import (
	"testing"

	"example.org/reportcsv"
)

func TestCell(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"plain", "plain"},
		{"red,blue", `"red,blue"`},
		{`say "hi"`, `"say ""hi"""`},
	} {
		if got := reportcsv.Cell(tc.input); got != tc.want {
			t.Errorf("Cell(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
