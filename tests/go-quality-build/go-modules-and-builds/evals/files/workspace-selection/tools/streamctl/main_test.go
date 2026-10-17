package main

import (
	"bytes"
	"testing"
)

func TestInspect(t *testing.T) {
	var out bytes.Buffer
	if err := inspect("sc1.b3JkZXJzLmV1.2s", &out); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "orders.eu\n100\n" {
		t.Fatalf("output = %q", got)
	}
	out.Reset()
	if err := inspect("broken", &out); err == nil || out.Len() != 0 {
		t.Fatalf("invalid cursor: output %q, error %v", out.String(), err)
	}
}
