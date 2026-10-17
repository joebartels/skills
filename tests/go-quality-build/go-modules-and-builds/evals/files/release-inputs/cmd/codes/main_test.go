package main

import (
	"bytes"
	"testing"
)

func TestRun(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		out  string
	}{
		{[]string{"invoice", "INV-123456"}, 0, "valid\n"},
		{[]string{"invoice", "INV-12345"}, 1, "invalid\n"},
		{[]string{"unknown", "INV-123456"}, 1, "invalid\n"},
	} {
		var stdout, stderr bytes.Buffer
		if got := run(tc.args, &stdout, &stderr); got != tc.code || stdout.String() != tc.out || stderr.Len() != 0 {
			t.Errorf("run(%q) = (%d, %q, %q)", tc.args, got, stdout.String(), stderr.String())
		}
	}
}
