package configpatch_test

import (
	"testing"

	"example.com/configpatch"
)

func TestApply(t *testing.T) {
	cfg := configpatch.Config{Retries: 2}
	if err := configpatch.Apply(&cfg, []byte(`{"enabled":true,"retries":4}`)); err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled || cfg.Retries != 4 {
		t.Fatalf("config = %#v", cfg)
	}
}
