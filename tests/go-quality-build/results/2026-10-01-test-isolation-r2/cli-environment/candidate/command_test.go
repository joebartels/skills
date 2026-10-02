package showcfg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestShowcfg(t *testing.T) {
	endpoint, endpointPresent := os.LookupEnv("INDEXER_ENDPOINT")
	mode, modePresent := os.LookupEnv("INDEXER_MODE")
	binary := filepath.Join(t.TempDir(), "showcfg")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/showcfg")
	build.WaitDelay = 2 * time.Second
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build showcfg: %v (context: %v)\n%s", err, ctx.Err(), output)
	}

	// The grouping Run waits for parallel children before parent assertions.
	// The parent's TempDir remains available through all descendants.
	t.Run("cases", func(t *testing.T) {
		for _, tc := range configurationCases() {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				command := exec.CommandContext(ctx, binary)
				command.WaitDelay = 2 * time.Second
				command.Env = childEnvironment(tc.endpoint, tc.mode)
				var stdout, stderr bytes.Buffer
				command.Stdout = &stdout
				command.Stderr = &stderr
				err := command.Run()
				if ctx.Err() != nil {
					t.Fatalf("showcfg deadline: %v; stdout=%q stderr=%q", ctx.Err(), stdout.String(), stderr.String())
				}
				if tc.errorKey != "" {
					var exitError *exec.ExitError
					if !errors.As(err, &exitError) || exitError.ExitCode() != 2 {
						t.Errorf("showcfg error = %v; want exit 2; stderr=%q", err, stderr.String())
					}
					if stdout.Len() != 0 {
						t.Errorf("failure stdout = %q; want empty", stdout.String())
					}
					if !strings.Contains(stderr.String(), tc.errorKey) {
						t.Errorf("failure stderr = %q; want diagnostic naming %s", stderr.String(), tc.errorKey)
					}
					return
				}
				if err != nil {
					t.Fatalf("showcfg: %v; stderr=%q", err, stderr.String())
				}
				if stderr.Len() != 0 {
					t.Errorf("success stderr = %q; want empty", stderr.String())
				}
				output := stdout.String()
				if !strings.HasSuffix(output, "\n") || strings.Count(output, "\n") != 1 {
					t.Errorf("stdout = %q; want one JSON value followed by one newline", output)
				}
				decoder := json.NewDecoder(strings.NewReader(output))
				var got map[string]string
				if err := decoder.Decode(&got); err != nil {
					t.Fatalf("decode stdout %q: %v", output, err)
				}
				if len(got) != 2 || got["endpoint"] != tc.want.Endpoint || got["mode"] != tc.want.Mode {
					t.Errorf("stdout JSON = %#v; want endpoint=%q and mode=%q", got, tc.want.Endpoint, tc.want.Mode)
				}
				var extra any
				if err := decoder.Decode(&extra); err != io.EOF {
					t.Errorf("after Config JSON: value=%#v error=%v; want EOF", extra, err)
				}
			})
		}
	})
	assertEnvironment(t, "INDEXER_ENDPOINT", envValue{endpoint, endpointPresent})
	assertEnvironment(t, "INDEXER_MODE", envValue{mode, modePresent})
}

func childEnvironment(endpoint, mode envValue) []string {
	environment := make([]string, 0, len(os.Environ())+2)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.EqualFold(key, "INDEXER_ENDPOINT") || strings.EqualFold(key, "INDEXER_MODE") {
			continue
		}
		environment = append(environment, entry)
	}
	if endpoint.present {
		environment = append(environment, "INDEXER_ENDPOINT="+endpoint.value)
	}
	if mode.present {
		environment = append(environment, "INDEXER_MODE="+mode.value)
	}
	return environment
}
