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
	wantErr  string
}

func configurationCases() []configurationCase {
	return []configurationCase{
		{name: "defaults", want: showcfg.Config{Endpoint: "http://127.0.0.1:9090", Mode: "read"}},
		{name: "explicit_https_write", endpoint: supplied("https://example.invalid/api"), mode: supplied("write"), want: showcfg.Config{Endpoint: "https://example.invalid/api", Mode: "write"}},
		{name: "explicit_http_read", endpoint: supplied("http://example.invalid:8080/api"), mode: supplied("read"), want: showcfg.Config{Endpoint: "http://example.invalid:8080/api", Mode: "read"}},
		{name: "endpoint_only", endpoint: supplied("https://example.invalid"), want: showcfg.Config{Endpoint: "https://example.invalid", Mode: "read"}},
		{name: "mode_only", mode: supplied("write"), want: showcfg.Config{Endpoint: "http://127.0.0.1:9090", Mode: "write"}},
		{name: "empty_endpoint", endpoint: supplied(""), wantErr: "INDEXER_ENDPOINT"},
		{name: "malformed_endpoint", endpoint: supplied("http://example.invalid/%zz"), wantErr: "INDEXER_ENDPOINT"},
		{name: "missing_scheme", endpoint: supplied("example.invalid/api"), wantErr: "INDEXER_ENDPOINT"},
		{name: "missing_host", endpoint: supplied("http:///api"), wantErr: "INDEXER_ENDPOINT"},
		{name: "unsupported_scheme", endpoint: supplied("ftp://example.invalid"), wantErr: "INDEXER_ENDPOINT"},
		{name: "empty_mode", mode: supplied(""), wantErr: "INDEXER_MODE"},
		{name: "unknown_mode", mode: supplied("delete"), wantErr: "INDEXER_MODE"},
		{name: "uppercase_mode", mode: supplied("READ"), wantErr: "INDEXER_MODE"},
		{name: "whitespace_mode", mode: supplied(" write "), wantErr: "INDEXER_MODE"},
	}
}

// Environment mutations must stay serial, including their ancestors.
func setConfigurationEnv(t *testing.T, endpoint, mode envValue) {
	t.Helper()
	for key, value := range map[string]envValue{"INDEXER_ENDPOINT": endpoint, "INDEXER_MODE": mode} {
		t.Setenv(key, value.value)
		if !value.present {
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("unset %s: %v", key, err)
			}
		}
	}
}

func currentConfigurationEnv() map[string]envValue {
	values := make(map[string]envValue, 2)
	for _, key := range []string{"INDEXER_ENDPOINT", "INDEXER_MODE"} {
		value, present := os.LookupEnv(key)
		values[key] = envValue{value: value, present: present}
	}
	return values
}

func assertConfigurationEnv(t *testing.T, want map[string]envValue) {
	t.Helper()
	for key, got := range currentConfigurationEnv() {
		if got != want[key] {
			t.Errorf("environment %s = %#v; want %#v", key, got, want[key])
		}
	}
}

func TestLoadFromEnv(t *testing.T) {
	for _, tc := range configurationCases() {
		t.Run(tc.name, func(t *testing.T) {
			setConfigurationEnv(t, tc.endpoint, tc.mode)
			before := currentConfigurationEnv()
			got, err := showcfg.LoadFromEnv()
			assertConfigurationEnv(t, before)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("LoadFromEnv error = %v; want diagnostic naming %s", err, tc.wantErr)
				}
				if got != (showcfg.Config{}) {
					t.Fatalf("LoadFromEnv = %#v on failure; want zero Config", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadFromEnv: %v", err)
			}
			if got != tc.want {
				t.Fatalf("LoadFromEnv = %#v; want %#v", got, tc.want)
			}
		})
	}
}

func TestConfigurationEnvironmentRestoration(t *testing.T) {
	for _, original := range []struct {
		name  string
		value envValue
	}{
		{name: "unset"},
		{name: "present_empty", value: supplied("")},
		{name: "present_value", value: supplied("parent sentinel")},
	} {
		t.Run(original.name, func(t *testing.T) {
			setConfigurationEnv(t, original.value, original.value)
			before := currentConfigurationEnv()
			t.Run("child", func(t *testing.T) {
				setConfigurationEnv(t, supplied("https://example.invalid"), supplied("write"))
				if _, err := showcfg.LoadFromEnv(); err != nil {
					t.Fatal(err)
				}
			})
			assertConfigurationEnv(t, before)
		})
	}
}
