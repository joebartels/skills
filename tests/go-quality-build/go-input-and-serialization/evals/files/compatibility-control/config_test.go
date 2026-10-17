package config_test

import (
	"testing"

	config "example.com/compatibleconfig"
)

func TestDecodeExistingConfig(t *testing.T) {
	cfg, err := config.Decode([]byte(`{"name":"worker","tags":["east"]}`))
	if err != nil || cfg.Name != "worker" || len(cfg.Tags) != 1 || cfg.Tags[0] != "east" {
		t.Fatalf("Decode = %+v, %v", cfg, err)
	}
}
