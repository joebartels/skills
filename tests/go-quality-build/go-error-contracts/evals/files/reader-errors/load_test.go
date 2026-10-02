package recordload

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	got, err := Load(strings.NewReader("a=1\nb=2\n"))
	if err != nil || len(got) != 2 || got[1].Key != "b" {
		t.Fatalf("%v %v", got, err)
	}
}
