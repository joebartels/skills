package showcfg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Keep Go's builder settings (including its ordinary default cache inputs)
// separate from application configuration. Do not require a GOCACHE override.
func commandEnvironment(endpoint, mode envValue) []string {
	env := make([]string, 0, len(os.Environ())+2)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key != "INDEXER_ENDPOINT" && key != "INDEXER_MODE" {
			env = append(env, entry)
		}
	}
	if endpoint.present {
		env = append(env, "INDEXER_ENDPOINT="+endpoint.value)
	}
	if mode.present {
		env = append(env, "INDEXER_MODE="+mode.value)
	}
	return env
}

func TestShowcfgCommand(t *testing.T) {
	// A command receives only its explicit application settings, even when
	// the parent holds invalid supplied values.
	setConfigurationEnv(t, supplied("invalid parent endpoint"), supplied("invalid parent mode"))
	parentEnv := currentConfigurationEnv()
	t.Cleanup(func() { assertConfigurationEnv(t, parentEnv) })

	binary := filepath.Join(t.TempDir(), "showcfg")
	buildCtx, cancelBuild := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancelBuild()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binary, "./cmd/showcfg")
	build.Env = commandEnvironment(envValue{}, envValue{})
	build.WaitDelay = 2 * time.Second
	output, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("build showcfg: %v (context: %v)\n%s", err, buildCtx.Err(), output)
	}

	for _, tc := range configurationCases() {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, binary)
			command.Env = commandEnvironment(tc.endpoint, tc.mode)
			command.WaitDelay = time.Second
			var stdout, stderr bytes.Buffer
			command.Stdout = &stdout
			command.Stderr = &stderr
			err := command.Run()
			if ctx.Err() != nil {
				t.Fatalf("showcfg exceeded deadline: %v; stdout=%q stderr=%q", ctx.Err(), stdout.String(), stderr.String())
			}
			assertConfigurationEnv(t, parentEnv)
			if tc.wantErr != "" {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() != 2 {
					t.Fatalf("showcfg error = %v; want exit 2; stderr=%q", err, stderr.String())
				}
				if stdout.Len() != 0 {
					t.Errorf("failure stdout = %q; want empty", stdout.String())
				}
				if !strings.Contains(stderr.String(), tc.wantErr) {
					t.Errorf("stderr = %q; want diagnostic naming %s", stderr.String(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("showcfg: %v; stderr=%q", err, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Errorf("success stderr = %q; want empty", stderr.String())
			}
			decoder := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
			var got map[string]string
			if err := decoder.Decode(&got); err != nil {
				t.Fatalf("decode stdout %q: %v", stdout.String(), err)
			}
			want := map[string]string{"endpoint": tc.want.Endpoint, "mode": tc.want.Mode}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("JSON = %#v; want %#v", got, want)
			}
			if remaining := stdout.Bytes()[decoder.InputOffset():]; !bytes.Equal(remaining, []byte("\n")) {
				t.Errorf("stdout after JSON = %q; want exactly one newline", remaining)
			}
		})
	}
}
