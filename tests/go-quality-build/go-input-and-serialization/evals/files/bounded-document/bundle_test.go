package bundlecfg_test

import (
	"strings"
	"testing"

	"example.com/bundlecfg"
)

func TestLoad(t *testing.T) {
	calls := 0
	err := bundlecfg.Load(strings.NewReader(`{"services":[{"name":"api","port":8080}]}`), func(b bundlecfg.Bundle) error {
		calls++
		if len(b.Services) != 1 || b.Services[0].Name != "api" || b.Services[0].Port != 8080 {
			t.Fatalf("bundle = %#v", b)
		}
		return nil
	})
	if err != nil || calls != 1 {
		t.Fatalf("Load: calls=%d, err=%v", calls, err)
	}
}
