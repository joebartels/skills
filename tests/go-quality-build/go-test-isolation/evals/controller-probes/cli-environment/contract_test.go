package showcfg_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/showcfg"
)

func TestEnvironmentRestored(t *testing.T) {
	t.Setenv("INDEXER_ENDPOINT", "https://parent.invalid/base")
	t.Setenv("INDEXER_MODE", "write")
	want := showcfg.Config{Endpoint: "https://parent.invalid/base", Mode: "write"}
	t.Run("temporary", func(t *testing.T) {
		t.Setenv("INDEXER_ENDPOINT", "http://child.invalid/api")
		t.Setenv("INDEXER_MODE", "read")
		got, err := showcfg.LoadFromEnv()
		if err != nil || got != (showcfg.Config{Endpoint: "http://child.invalid/api", Mode: "read"}) {
			t.Fatalf("temporary = %#v, %v", got, err)
		}
	})
	if got, err := showcfg.LoadFromEnv(); err != nil || got != want {
		t.Fatalf("parent = %#v, %v; want %#v", got, err, want)
	}
}

func TestChildConfiguration(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "showcfg")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/showcfg")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	for _, tc := range []struct {
		name string
		env  []string
		want showcfg.Config
		code int
	}{
		{"unset", nil, showcfg.Config{Endpoint: "http://127.0.0.1:9090", Mode: "read"}, 0},
		{"explicit", []string{"INDEXER_ENDPOINT=https://child.invalid/item?q=x", "INDEXER_MODE=write"}, showcfg.Config{Endpoint: "https://child.invalid/item?q=x", Mode: "write"}, 0},
		{"invalid", []string{"INDEXER_ENDPOINT=ftp://child.invalid", "INDEXER_MODE=read"}, showcfg.Config{}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary)
			for _, entry := range os.Environ() {
				if !strings.HasPrefix(entry, "INDEXER_ENDPOINT=") && !strings.HasPrefix(entry, "INDEXER_MODE=") {
					cmd.Env = append(cmd.Env, entry)
				}
			}
			cmd.Env = append(cmd.Env, tc.env...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			if ctx.Err() != nil {
				t.Fatal(ctx.Err())
			}
			if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != tc.code {
				t.Fatalf("process = %v, %v; want exit %d", cmd.ProcessState, err, tc.code)
			}
			if tc.code != 0 {
				if stdout.Len() != 0 || strings.TrimSpace(stderr.String()) == "" {
					t.Fatalf("failure stdout=%q stderr=%q", stdout.String(), stderr.String())
				}
				return
			}
			want, _ := json.Marshal(tc.want)
			if stdout.String() != string(want)+"\n" || stderr.Len() != 0 {
				t.Fatalf("stdout=%q stderr=%q want %q", stdout.String(), stderr.String(), string(want)+"\n")
			}
		})
	}
}
