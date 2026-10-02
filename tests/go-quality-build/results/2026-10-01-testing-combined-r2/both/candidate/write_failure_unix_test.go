//go:build unix

package indexer_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"example.com/indexer"
)

func TestPublicationWriteFailure(t *testing.T) {
	if operation := os.Getenv("INDEXER_WRITE_FAILURE_HELPER"); operation != "" {
		// This limit belongs only to this terminable child, never to the test host.
		signal.Ignore(syscall.SIGXFSZ)
		if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 1024, Max: 1024}); err != nil {
			t.Fatal(err)
		}
		root := os.Getenv("INDEXER_WRITE_FAILURE_ROOT")
		huge := strings.Repeat("z", 8192)
		var err error
		switch operation {
		case "put":
			err = indexer.Open(root).Put("beta", huge)
		case "apply":
			err = indexer.Open(root).ApplyCSV(strings.NewReader("alpha,prefix\nbeta,accepted\nbeta," + huge + "\ngamma,later\n"))
		case "refresh":
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Body: &observedBody{reader: strings.NewReader(`[{"key":"beta","text":"` + huge + `"}]`)}}, nil
			})}
			err = indexer.Refresh(context.Background(), client, "http://example.test", filepath.Join(root, "beta"))
		default:
			t.Fatalf("unknown helper operation %s", operation)
		}
		if err == nil {
			t.Fatal("expected write failure after partial temporary-file write")
		}
		if !strings.Contains(err.Error(), "file too large") {
			t.Fatalf("wrong failure boundary: %v", err)
		}
		return
	}
	for _, operation := range []string{"put", "apply", "refresh"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			path := filepath.Join(root, "beta")
			prior := "exact old bytes\x00\n"
			if err := os.WriteFile(path, []byte(prior), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPublicationWriteFailure$", "-test.timeout=4s")
			cmd.Env = childEnvironment()
			cmd.Env = append(cmd.Env, "INDEXER_WRITE_FAILURE_HELPER="+operation, "INDEXER_WRITE_FAILURE_ROOT="+root)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("write-failure child = %v, deadline=%v\n%s", err, ctx.Err(), output)
			}
			if operation == "apply" {
				assertFile(t, path, "accepted")
			} else {
				assertFile(t, path, prior)
			}
			if operation == "apply" {
				assertText(t, indexer.Open(root), "alpha", "prefix")
				assertMissing(t, indexer.Open(root), "gamma")
			}
			assertNoTemps(t, root)
		})
	}
}

func childEnvironment() []string {
	result := []string{}
	for _, name := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOTOOLCHAIN"} {
		if value, ok := os.LookupEnv(name); ok {
			result = append(result, fmt.Sprintf("%s=%s", name, value))
		}
	}
	return result
}
