package showcfg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"example.com/showcfg"
)

type configCase struct {
	name     string
	env      map[string]string
	want     showcfg.Config
	errorKey string
}

func configurationCases() []configCase {
	return []configCase{
		{
			name: "defaults",
			want: showcfg.Config{Endpoint: "http://127.0.0.1:9090", Mode: "read"},
		},
		{
			name: "explicit_http_read",
			env:  map[string]string{"INDEXER_ENDPOINT": "http://example.invalid/api", "INDEXER_MODE": "read"},
			want: showcfg.Config{Endpoint: "http://example.invalid/api", Mode: "read"},
		},
		{
			name: "explicit_https_write",
			env:  map[string]string{"INDEXER_ENDPOINT": "https://example.invalid/api?limit=10", "INDEXER_MODE": "write"},
			want: showcfg.Config{Endpoint: "https://example.invalid/api?limit=10", Mode: "write"},
		},
		{
			name: "endpoint_only",
			env:  map[string]string{"INDEXER_ENDPOINT": "https://example.invalid"},
			want: showcfg.Config{Endpoint: "https://example.invalid", Mode: "read"},
		},
		{
			name: "mode_only",
			env:  map[string]string{"INDEXER_MODE": "write"},
			want: showcfg.Config{Endpoint: "http://127.0.0.1:9090", Mode: "write"},
		},
		{
			name:     "empty_endpoint",
			env:      map[string]string{"INDEXER_ENDPOINT": ""},
			errorKey: "INDEXER_ENDPOINT",
		},
		{
			name:     "missing_scheme",
			env:      map[string]string{"INDEXER_ENDPOINT": "example.invalid/api"},
			errorKey: "INDEXER_ENDPOINT",
		},
		{
			name:     "unsupported_scheme",
			env:      map[string]string{"INDEXER_ENDPOINT": "ftp://example.invalid/api"},
			errorKey: "INDEXER_ENDPOINT",
		},
		{
			name:     "missing_host",
			env:      map[string]string{"INDEXER_ENDPOINT": "http:///api"},
			errorKey: "INDEXER_ENDPOINT",
		},
		{
			name:     "malformed_endpoint",
			env:      map[string]string{"INDEXER_ENDPOINT": "https://%zz"},
			errorKey: "INDEXER_ENDPOINT",
		},
		{
			name:     "empty_mode",
			env:      map[string]string{"INDEXER_MODE": ""},
			errorKey: "INDEXER_MODE",
		},
		{
			name:     "unknown_mode",
			env:      map[string]string{"INDEXER_MODE": "delete"},
			errorKey: "INDEXER_MODE",
		},
		{
			name:     "capitalized_mode",
			env:      map[string]string{"INDEXER_MODE": "Read"},
			errorKey: "INDEXER_MODE",
		},
		{
			name:     "whitespace_mode",
			env:      map[string]string{"INDEXER_MODE": " read "},
			errorKey: "INDEXER_MODE",
		},
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Environment mutation and all ancestors stay serial.
	original := currentEnvironment()
	for _, tc := range configurationCases() {
		t.Run(tc.name, func(t *testing.T) {
			setEnvironment(t, tc.env)
			before := currentEnvironment()
			got, err := showcfg.LoadFromEnv()
			assertEnvironment(t, before)
			if tc.errorKey != "" {
				if got != (showcfg.Config{}) {
					t.Errorf("LoadFromEnv = %#v; want zero Config on error", got)
				}
				if err == nil {
					t.Fatal("LoadFromEnv succeeded; want an error")
				}
				assertDiagnostic(t, err.Error(), tc)
				return
			}
			if err != nil {
				t.Fatalf("LoadFromEnv: %v", err)
			}
			if got != tc.want {
				t.Errorf("LoadFromEnv = %#v; want %#v", got, tc.want)
			}
		})
		// Subtest cleanup must restore both the value and its presence.
		assertEnvironment(t, original)
	}
}

func TestEnvironmentRestoration(t *testing.T) {
	parents := []struct {
		name string
		env  map[string]string
	}{
		{name: "unset"},
		{name: "present_empty", env: map[string]string{"INDEXER_ENDPOINT": "", "INDEXER_MODE": ""}},
		{name: "present_values", env: map[string]string{"INDEXER_ENDPOINT": "https://parent.invalid", "INDEXER_MODE": "write"}},
	}
	for _, parent := range parents {
		t.Run(parent.name, func(t *testing.T) {
			setEnvironment(t, parent.env)
			before := currentEnvironment()
			for _, tc := range configurationCases() {
				t.Run(tc.name, func(t *testing.T) {
					setEnvironment(t, tc.env)
					_, _ = showcfg.LoadFromEnv()
				})
				assertEnvironment(t, before)
			}
		})
	}
}

func TestShowcfgCommand(t *testing.T) {
	// Poison the parent to expose accidental inheritance in defaults and
	// partially supplied cases. t.Setenv owns restoration for this test.
	t.Setenv("INDEXER_ENDPOINT", "parent://ignored")
	t.Setenv("INDEXER_MODE", "parent-mode")
	parent := currentEnvironment()
	binary := filepath.Join(t.TempDir(), "showcfg")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/showcfg")
	build.Env = commandEnvironment(nil)
	build.WaitDelay = 5 * time.Second
	output, err := build.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("build command deadline: %v; output: %s", ctx.Err(), output)
	}
	if err != nil {
		t.Fatalf("build command: %v; output: %s", err, output)
	}
	for _, tc := range configurationCases() {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary)
			cmd.Env = commandEnvironment(tc.env)
			cmd.WaitDelay = 5 * time.Second
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			if ctx.Err() != nil {
				t.Fatalf("command deadline: %v; stdout: %q; stderr: %q", ctx.Err(), stdout.String(), stderr.String())
			}
			if tc.errorKey != "" {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() != 2 {
					t.Fatalf("command error = %v; want exit 2; stderr: %q", err, stderr.String())
				}
				if stdout.Len() != 0 {
					t.Errorf("stdout = %q; want empty on failure", stdout.String())
				}
				assertDiagnostic(t, stderr.String(), tc)
				return
			}
			if err != nil {
				t.Fatalf("command: %v; stderr: %q", err, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Errorf("stderr = %q; want empty on success", stderr.String())
			}
			assertConfigJSON(t, stdout.Bytes(), tc.want)
		})
		assertEnvironment(t, parent)
	}
}

func setEnvironment(t *testing.T, values map[string]string) {
	t.Helper()
	for _, key := range []string{"INDEXER_ENDPOINT", "INDEXER_MODE"} {
		value, present := values[key]
		// Register restoration before unsetting, including on a setup failure.
		t.Setenv(key, value)
		if !present {
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("unset %s: %v", key, err)
			}
		}
	}
}

type envValue struct {
	value   string
	present bool
}

func currentEnvironment() map[string]envValue {
	values := make(map[string]envValue)
	for _, key := range []string{"INDEXER_ENDPOINT", "INDEXER_MODE"} {
		value, present := os.LookupEnv(key)
		values[key] = envValue{value: value, present: present}
	}
	return values
}

func assertEnvironment(t *testing.T, want map[string]envValue) {
	t.Helper()
	for key, expected := range want {
		value, present := os.LookupEnv(key)
		if got := (envValue{value: value, present: present}); got != expected {
			t.Errorf("environment %s = %#v; want %#v", key, got, expected)
		}
	}
}

func commandEnvironment(values map[string]string) []string {
	env := make([]string, 0, len(os.Environ())+len(values))
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		// Environment keys are case-insensitive on Windows.
		if strings.EqualFold(key, "INDEXER_ENDPOINT") || strings.EqualFold(key, "INDEXER_MODE") {
			continue
		}
		env = append(env, entry)
	}
	for key, value := range values {
		env = append(env, key+"="+value)
	}
	return env
}

func assertDiagnostic(t *testing.T, diagnostic string, tc configCase) {
	t.Helper()
	if !strings.Contains(diagnostic, tc.errorKey) {
		t.Errorf("diagnostic = %q; want offending variable %s", diagnostic, tc.errorKey)
	}
	if value := tc.env[tc.errorKey]; value != "" && !strings.Contains(diagnostic, value) {
		t.Errorf("diagnostic = %q; want offending value %q", diagnostic, value)
	}
}

func assertConfigJSON(t *testing.T, output []byte, want showcfg.Config) {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(output))
	var got map[string]string
	if err := decoder.Decode(&got); err != nil {
		t.Fatalf("decode command stdout %q: %v", output, err)
	}
	endpoint, hasEndpoint := got["endpoint"]
	mode, hasMode := got["mode"]
	if len(got) != 2 || !hasEndpoint || !hasMode || endpoint != want.Endpoint || mode != want.Mode {
		t.Errorf("command JSON = %#v; want lowercase endpoint %q and mode %q only", got, want.Endpoint, want.Mode)
	}
	if trailing := string(output[decoder.InputOffset():]); trailing != "\n" {
		t.Errorf("bytes after Config JSON = %q; want exactly one newline", trailing)
	}
}
