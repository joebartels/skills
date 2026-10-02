package indexer_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/indexer"
)

// Function assignments exercise the consumer-visible signatures.
var (
	_ func(string) *indexer.Index                                                           = indexer.Open
	_ func(*indexer.Index, string, string) error                                            = (*indexer.Index).Put
	_ func(*indexer.Index, string) (string, error)                                          = (*indexer.Index).Get
	_ func(*indexer.Index, io.Reader) error                                                 = (*indexer.Index).ApplyCSV
	_ func(context.Context, *http.Client, string, string) error                             = indexer.Refresh
	_ func(context.Context, time.Duration, func(context.Context) error, func() error) error = indexer.Serve
)

func TestExistingPutAndReplace(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	for _, text := range []string{"first", "x", "", " \n\t", "\x00arbitrary"} {
		if err := i.Put("alpha", text); err != nil {
			t.Fatal(err)
		}
		assertText(t, i, "alpha", text)
	}
	for _, key := range []string{"", "Alpha", "a1", "../escape", "é"} {
		if err := i.Put(key, "bad"); !errors.Is(err, indexer.ErrInvalidRecord) {
			t.Fatalf("Put(%q) = %v; want ErrInvalidRecord", key, err)
		}
		if _, err := i.Get(key); !errors.Is(err, indexer.ErrInvalidRecord) {
			t.Fatalf("Get(%q) = %v; want ErrInvalidRecord", key, err)
		}
	}
}

func TestIndependentRoots(t *testing.T) {
	t.Parallel()
	first := indexer.Open(filepath.Join(t.TempDir(), "first"))
	second := indexer.Open(filepath.Join(t.TempDir(), "second"))
	if err := first.Put("alpha", "first root"); err != nil {
		t.Fatal(err)
	}
	if err := second.Put("alpha", "second root"); err != nil {
		t.Fatal(err)
	}
	assertText(t, first, "alpha", "first root")
	assertText(t, second, "alpha", "second root")
}

func TestApplyCSVQuotedAndDuplicateRecords(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	input := "alpha,old\nbeta,\"  commas, and\nnewlines  \"\nalpha,replacement\n"
	if err := i.ApplyCSV(strings.NewReader(input)); err != nil {
		t.Fatal(err)
	}
	assertText(t, i, "alpha", "replacement")
	assertText(t, i, "beta", "  commas, and\nnewlines  ")
}

func TestApplyCSVRejectedReplacementKeepsPrefix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		row     string
		invalid bool
	}{
		{"blank", "alpha,\" \t\u2003 \"\n", true},
		{"uppercase", "Alpha,text\n", true},
		{"traversal", "../escape,text\n", true},
		{"missing-field", "alpha\n", true},
		{"extra-field", "alpha,text,extra\n", true},
		{"parse", "alpha,\"unterminated\n", false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			i := indexer.Open(t.TempDir())
			input := "alpha,accepted\nbeta,also accepted\n" + test.row + "gamma,later\n"
			err := i.ApplyCSV(strings.NewReader(input))
			if err == nil || (test.invalid && !errors.Is(err, indexer.ErrInvalidRecord)) {
				t.Fatalf("ApplyCSV = %v; want rejection (invalid=%v)", err, test.invalid)
			}
			assertText(t, i, "alpha", "accepted")
			assertText(t, i, "beta", "also accepted")
			if _, err := i.Get("gamma"); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("later row exists: %v", err)
			}
		})
	}
}

func TestApplyCSVWriteFailureKeepsPrefix(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "beta"), 0700); err != nil {
		t.Fatal(err)
	}
	i := indexer.Open(root)
	if err := i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,cannot replace directory\ngamma,later\n")); err == nil {
		t.Fatal("wanted publication failure")
	}
	assertText(t, i, "alpha", "accepted")
	if _, err := i.Get("gamma"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("later row exists: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %v; want only alpha and beta, without temporary files", entries)
	}
}

func TestApplyCSVReadFailureKeepsPrefix(t *testing.T) {
	t.Parallel()
	readErr := errors.New("reader stopped")
	i := indexer.Open(t.TempDir())
	r := io.MultiReader(strings.NewReader("alpha,accepted\n"), errorReader{readErr})
	if err := i.ApplyCSV(r); !errors.Is(err, readErr) {
		t.Fatalf("ApplyCSV = %v; want read error", err)
	}
	assertText(t, i, "alpha", "accepted")
}

func TestPutWriteFailureKeepsOld(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	i := indexer.Open(root)
	if err := i.Put("alpha", "exact old bytes\x00\n"); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(root, 0700) })
	probe, err := os.CreateTemp(root, ".permission-probe-*")
	if err == nil {
		probe.Close()
		os.Remove(probe.Name())
		t.Skip("filesystem privileges bypass directory write permission")
	}
	if err := i.Put("alpha", "new"); err == nil {
		t.Fatal("wanted write failure")
	}
	assertText(t, i, "alpha", "exact old bytes\x00\n")
}

func assertText(t *testing.T, i *indexer.Index, key, want string) {
	t.Helper()
	got, err := i.Get(key)
	if err != nil || got != want {
		t.Fatalf("Get(%q) = %q, %v; want %q", key, got, err, want)
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }
