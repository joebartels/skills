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

func supplied(value string) envValue {
	return envValue{value: value, present: true}
}

type configurationCase struct {
	name     string
	endpoint envValue
	mode     envValue
	want     showcfg.Config
	errKey   string
}

func configurationCases() []configurationCase {
	return []configurationCase{
		{
			name: "defaults",
			want: showcfg.Config{Endpoint: "http://127.0.0.1:9090", Mode: "read"},
		},
		{
			name:     "http endpoint with default mode",
			endpoint: supplied("http://example.invalid:8080/api"),
			want:     showcfg.Config{Endpoint: "http://example.invalid:8080/api", Mode: "read"},
		},
		{
			name: "explicit read with default endpoint",
			mode: supplied("read"),
			want: showcfg.Config{Endpoint: "http://127.0.0.1:9090", Mode: "read"},
		},
		{
			name: "explicit write with default endpoint",
			mode: supplied("write"),
			want: showcfg.Config{Endpoint: "http://127.0.0.1:9090", Mode: "write"},
		},
		{
			name:     "https endpoint and write",
			endpoint: supplied("https://example.invalid/api?limit=10"),
			mode:     supplied("write"),
			want:     showcfg.Config{Endpoint: "https://example.invalid/api?limit=10", Mode: "write"},
		},
		{name: "empty endpoint", endpoint: supplied(""), errKey: "INDEXER_ENDPOINT"},
		{name: "missing scheme", endpoint: supplied("example.invalid/api"), errKey: "INDEXER_ENDPOINT"},
		{name: "unsupported scheme", endpoint: supplied("ftp://example.invalid/api"), errKey: "INDEXER_ENDPOINT"},
		{name: "missing host", endpoint: supplied("https:///api"), errKey: "INDEXER_ENDPOINT"},
		{name: "malformed endpoint", endpoint: supplied("http://%/api"), errKey: "INDEXER_ENDPOINT"},
		{name: "empty mode", mode: supplied(""), errKey: "INDEXER_MODE"},
		{name: "unknown mode", mode: supplied("delete"), errKey: "INDEXER_MODE"},
		{name: "wrong case mode", mode: supplied("READ"), errKey: "INDEXER_MODE"},
		{name: "whitespace mode", mode: supplied("write "), errKey: "INDEXER_MODE"},
	}
}

func lookupEnv(key string) envValue {
	value, present := os.LookupEnv(key)
	return envValue{value: value, present: present}
}

func setEnv(t *testing.T, key string, value envValue) {
	t.Helper()
	// Setenv registers restoration of both the old value and its presence.
	// Register that cleanup even when this case needs the variable unset.
	t.Setenv(key, value.value)
	if !value.present {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%q): %v", key, err)
		}
	}
}

func TestLoadFromEnv(t *testing.T) {
	// These subtests change process state and must remain serial.
	endpointBefore := lookupEnv("INDEXER_ENDPOINT")
	modeBefore := lookupEnv("INDEXER_MODE")
	for _, tc := range configurationCases() {
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, "INDEXER_ENDPOINT", tc.endpoint)
			setEnv(t, "INDEXER_MODE", tc.mode)
			got, err := showcfg.LoadFromEnv()
			if tc.errKey != "" {
				if err == nil {
					t.Fatalf("LoadFromEnv = %#v, nil; want error identifying %s", got, tc.errKey)
				}
				if !strings.Contains(err.Error(), tc.errKey) {
					t.Errorf("error %q does not identify %s", err, tc.errKey)
				}
				if got != (showcfg.Config{}) {
					t.Errorf("LoadFromEnv returned %#v on failure; want zero Config", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadFromEnv: %v", err)
			}
			if got != tc.want {
				t.Errorf("LoadFromEnv = %#v; want %#v", got, tc.want)
			}
		})
		// t.Run returns after the subtest's cleanup has run.
		if got := lookupEnv("INDEXER_ENDPOINT"); got != endpointBefore {
			t.Fatalf("endpoint environment changed after %q: %#v; want %#v", tc.name, got, endpointBefore)
		}
		if got := lookupEnv("INDEXER_MODE"); got != modeBefore {
			t.Fatalf("mode environment changed after %q: %#v; want %#v", tc.name, got, modeBefore)
		}
	}
}

func TestEnvironmentRestoration(t *testing.T) {
	for _, parent := range []struct {
		name     string
		endpoint envValue
		mode     envValue
	}{
		{name: "unset"},
		{name: "empty", endpoint: supplied(""), mode: supplied("")},
		{name: "supplied", endpoint: supplied("https://parent.invalid"), mode: supplied("write")},
	} {
		t.Run(parent.name, func(t *testing.T) {
			setEnv(t, "INDEXER_ENDPOINT", parent.endpoint)
			setEnv(t, "INDEXER_MODE", parent.mode)
			for _, child := range []struct {
				name     string
				endpoint envValue
				mode     envValue
			}{
				{name: "unset"},
				{name: "supplied", endpoint: supplied("https://child.invalid"), mode: supplied("read")},
			} {
				t.Run(child.name, func(t *testing.T) {
					setEnv(t, "INDEXER_ENDPOINT", child.endpoint)
					setEnv(t, "INDEXER_MODE", child.mode)
					if _, err := showcfg.LoadFromEnv(); err != nil {
						t.Fatalf("LoadFromEnv: %v", err)
					}
				})
				if got := lookupEnv("INDEXER_ENDPOINT"); got != parent.endpoint {
					t.Errorf("endpoint after child %q = %#v; want %#v", child.name, got, parent.endpoint)
				}
				if got := lookupEnv("INDEXER_MODE"); got != parent.mode {
					t.Errorf("mode after child %q = %#v; want %#v", child.name, got, parent.mode)
				}
			}
		})
	}
}
