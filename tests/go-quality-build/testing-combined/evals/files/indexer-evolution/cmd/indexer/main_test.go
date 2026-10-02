package main

import "testing"

func TestExistingPut(t *testing.T) {
	if err := run([]string{"put", t.TempDir(), "alpha", "one"}); err != nil {
		t.Fatal(err)
	}
	if err := run(nil); err == nil {
		t.Fatal("missing usage rejection")
	}
}
