package showcfg_test

import (
	"os"
	"strings"
	"testing"

	"example.com/showcfg"
)

type envValue struct {
	value   string
	present bool
}

type configCase struct {
	name     string
	endpoint envValue
	mode     envValue
	want     showcfg.Config
	errorKey string
}

func configurationCases() []configCase {
	return []configCase{
		{name: "defaults", want: showcfg.Config{Endpoint: "http://127.0.0.1:9090", Mode: "read"}},
		{name: "explicit_http_read", endpoint: envValue{"http://example.invalid:8080/api", true}, mode: envValue{"read", true}, want: showcfg.Config{Endpoint: "http://example.invalid:8080/api", Mode: "read"}},
		{name: "explicit_https_write", endpoint: envValue{"https://example.invalid/api", true}, mode: envValue{"write", true}, want: showcfg.Config{Endpoint: "https://example.invalid/api", Mode: "write"}},
		{name: "endpoint_only", endpoint: envValue{"https://example.invalid/api", true}, want: showcfg.Config{Endpoint: "https://example.invalid/api", Mode: "read"}},
		{name: "mode_only", mode: envValue{"write", true}, want: showcfg.Config{Endpoint: "http://127.0.0.1:9090", Mode: "write"}},
		{name: "empty_endpoint", endpoint: envValue{"", true}, errorKey: "INDEXER_ENDPOINT"},
		{name: "unsupported_scheme", endpoint: envValue{"ftp://example.invalid/api", true}, errorKey: "INDEXER_ENDPOINT"},
		{name: "missing_host", endpoint: envValue{"https:///api", true}, errorKey: "INDEXER_ENDPOINT"},
		{name: "relative_endpoint", endpoint: envValue{"/api", true}, errorKey: "INDEXER_ENDPOINT"},
		{name: "malformed_endpoint", endpoint: envValue{"https://example.invalid/%zz", true}, errorKey: "INDEXER_ENDPOINT"},
		{name: "empty_mode", mode: envValue{"", true}, errorKey: "INDEXER_MODE"},
		{name: "unsupported_mode", mode: envValue{"delete", true}, errorKey: "INDEXER_MODE"},
		{name: "uppercase_mode", mode: envValue{"READ", true}, errorKey: "INDEXER_MODE"},
		{name: "whitespace_mode", mode: envValue{" write ", true}, errorKey: "INDEXER_MODE"},
		{name: "valid_endpoint_invalid_mode", endpoint: envValue{"https://example.invalid/api", true}, mode: envValue{"invalid", true}, errorKey: "INDEXER_MODE"},
	}
}

// These tests stay serial because LoadFromEnv reads process-wide state.
func TestLoadFromEnv(t *testing.T) {
	for _, tc := range configurationCases() {
		t.Run(tc.name, func(t *testing.T) {
			setEnvironment(t, "INDEXER_ENDPOINT", tc.endpoint)
			setEnvironment(t, "INDEXER_MODE", tc.mode)
			got, err := showcfg.LoadFromEnv()
			if tc.errorKey != "" {
				if err == nil || !strings.Contains(err.Error(), tc.errorKey) {
					t.Errorf("LoadFromEnv error = %v; want diagnostic naming %s", err, tc.errorKey)
				}
				if got != (showcfg.Config{}) {
					t.Errorf("LoadFromEnv on failure = %#v; want zero Config", got)
				}
			} else if err != nil || got != tc.want {
				t.Errorf("LoadFromEnv = %#v, %v; want %#v, nil", got, err, tc.want)
			}
			assertEnvironment(t, "INDEXER_ENDPOINT", tc.endpoint)
			assertEnvironment(t, "INDEXER_MODE", tc.mode)
		})
	}
}

func TestEnvironmentRestored(t *testing.T) {
	for _, tc := range []struct {
		name  string
		state envValue
	}{
		{name: "unset"},
		{name: "present_empty", state: envValue{"", true}},
		{name: "present_value", state: envValue{"parent-sentinel", true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			setEnvironment(t, "INDEXER_ENDPOINT", tc.state)
			setEnvironment(t, "INDEXER_MODE", tc.state)
			for _, child := range []configCase{
				{name: "unset", want: showcfg.Config{Endpoint: "http://127.0.0.1:9090", Mode: "read"}},
				{name: "supplied", endpoint: envValue{"https://example.invalid/api", true}, mode: envValue{"write", true}, want: showcfg.Config{Endpoint: "https://example.invalid/api", Mode: "write"}},
				{name: "malformed", mode: envValue{"", true}, errorKey: "INDEXER_MODE"},
			} {
				t.Run(child.name, func(t *testing.T) {
					setEnvironment(t, "INDEXER_ENDPOINT", child.endpoint)
					setEnvironment(t, "INDEXER_MODE", child.mode)
					got, err := showcfg.LoadFromEnv()
					if child.errorKey != "" {
						if err == nil || got != (showcfg.Config{}) {
							t.Errorf("LoadFromEnv = %#v, %v; want zero Config and error", got, err)
						}
						return
					}
					if err != nil || got != child.want {
						t.Errorf("LoadFromEnv = %#v, %v; want %#v, nil", got, err, child.want)
					}
				})
				assertEnvironment(t, "INDEXER_ENDPOINT", tc.state)
				assertEnvironment(t, "INDEXER_MODE", tc.state)
			}
		})
	}
}

func setEnvironment(t *testing.T, key string, value envValue) {
	t.Helper()
	// Setenv registers restoration of both presence and value before Unsetenv.
	if err := os.Setenv(key, value.value); err != nil { t.Fatal(err) }
	if !value.present {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}
}

func assertEnvironment(t *testing.T, key string, want envValue) {
	t.Helper()
	value, present := os.LookupEnv(key)
	if got := (envValue{value, present}); got != want {
		t.Errorf("%s = %#v; want %#v", key, got, want)
	}
}
