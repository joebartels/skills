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
)

func commandEnv(endpoint, mode envValue) []string {
	env := make([]string, 0, len(os.Environ())+2)
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key == "INDEXER_ENDPOINT" || key == "INDEXER_MODE" {
			continue
		}
		if runtime.GOOS == "windows" && (strings.EqualFold(key, "INDEXER_ENDPOINT") || strings.EqualFold(key, "INDEXER_MODE")) {
			continue
		}
		env = append(env, entry)
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
	name := "showcfg"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	buildCtx, cancelBuild := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancelBuild()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binary, "./cmd/showcfg")
	build.WaitDelay = time.Second
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build showcfg: %v (context: %v)\n%s", err, buildCtx.Err(), output)
	}

	// A deliberately invalid parent proves defaults do not inherit the shell.
	t.Setenv("INDEXER_ENDPOINT", "invalid-parent-endpoint")
	t.Setenv("INDEXER_MODE", "invalid-parent-mode")
	endpointBefore := lookupEnv("INDEXER_ENDPOINT")
	modeBefore := lookupEnv("INDEXER_MODE")
	for _, tc := range configurationCases() {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary)
			cmd.WaitDelay = time.Second
			cmd.Env = commandEnv(tc.endpoint, tc.mode)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			if ctx.Err() != nil {
				t.Fatalf("showcfg exceeded deadline: %v; stdout=%q stderr=%q", ctx.Err(), stdout.String(), stderr.String())
			}
			if tc.errKey != "" {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("showcfg error = %v; want exit 2", err)
				}
				if exitErr.ExitCode() != 2 {
					t.Errorf("exit code = %d; want 2", exitErr.ExitCode())
				}
				if stdout.Len() != 0 {
					t.Errorf("failure stdout = %q; want empty", stdout.String())
				}
				if !strings.Contains(stderr.String(), tc.errKey) {
					t.Errorf("stderr = %q; want diagnostic identifying %s", stderr.String(), tc.errKey)
				}
				return
			}
			if err != nil {
				t.Fatalf("showcfg: %v; stderr=%q", err, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Errorf("success stderr = %q; want empty", stderr.String())
			}
			output := stdout.Bytes()
			if len(output) == 0 || output[len(output)-1] != '\n' {
				t.Fatalf("stdout = %q; want JSON followed by newline", output)
			}
			decoder := json.NewDecoder(bytes.NewReader(output))
			var got map[string]string
			if err := decoder.Decode(&got); err != nil {
				t.Fatalf("decode stdout %q: %v", output, err)
			}
			if len(got) != 2 || got["endpoint"] != tc.want.Endpoint || got["mode"] != tc.want.Mode {
				t.Errorf("JSON = %#v; want lowercase endpoint=%q mode=%q", got, tc.want.Endpoint, tc.want.Mode)
			}
			if trailing := output[decoder.InputOffset():]; !bytes.Equal(trailing, []byte("\n")) {
				t.Errorf("stdout after first JSON value = %q; want exactly one newline", trailing)
			}
		})
		if got := lookupEnv("INDEXER_ENDPOINT"); got != endpointBefore {
			t.Fatalf("child changed endpoint environment after %q: %#v; want %#v", tc.name, got, endpointBefore)
		}
		if got := lookupEnv("INDEXER_MODE"); got != modeBefore {
			t.Fatalf("child changed mode environment after %q: %#v; want %#v", tc.name, got, modeBefore)
		}
	}
}
