package roster

import (
	"reflect"
	"testing"
)

func TestSummarize(t *testing.T) {
	got := Summarize([]Entry{{Tenant: "a", Name: "Zed"}, {Tenant: "b", Name: "Ada"}, {Tenant: "a", Name: "Zed"}, {Tenant: "b", Name: ""}})
	if got.Total != 4 || !reflect.DeepEqual(got.Names, []string{"Ada", "Zed"}) {
		t.Fatalf("got %#v", got)
	}
	if empty := Summarize(nil); empty.Total != 0 || empty.Names == nil || len(empty.Names) != 0 {
		t.Fatalf("empty report %#v", empty)
	}
}
