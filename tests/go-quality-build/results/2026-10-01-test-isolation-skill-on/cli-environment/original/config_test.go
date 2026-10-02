package showcfg_test

import (
	"testing"

	"example.com/showcfg"
)

func TestExplicitConfiguration(t *testing.T) {
	t.Setenv("INDEXER_ENDPOINT", "https://example.invalid/api")
	t.Setenv("INDEXER_MODE", "write")
	got, err := showcfg.LoadFromEnv()
	want := showcfg.Config{Endpoint: "https://example.invalid/api", Mode: "write"}
	if err != nil || got != want {
		t.Fatalf("LoadFromEnv = %#v, %v; want %#v", got, err, want)
	}
}
