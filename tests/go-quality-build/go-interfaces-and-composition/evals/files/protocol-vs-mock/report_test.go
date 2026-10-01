package reportkit_test

import (
	reportkit "example.com/protocol-vs-mock"
	"testing"
)

func TestJSONReport(t *testing.T) {
	r := reportkit.Report{Encoder: reportkit.JSON{}}
	got, err := r.Render([]reportkit.Record{{Name: "total", Value: "7"}})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "[{\"name\":\"total\",\"value\":\"7\"}]\n" {
		t.Fatalf("report = %q", got)
	}
}
