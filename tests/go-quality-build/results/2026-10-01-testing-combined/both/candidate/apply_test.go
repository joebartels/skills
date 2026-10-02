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

// Assignments exercise consumer function types as well as ordinary calls.
var (
	_ func(string) *indexer.Index                                                           = indexer.Open
	_ func(*indexer.Index, string, string) error                                            = (*indexer.Index).Put
	_ func(*indexer.Index, string) (string, error)                                          = (*indexer.Index).Get
	_ func(*indexer.Index, io.Reader) error                                                 = (*indexer.Index).ApplyCSV
	_ func(context.Context, *http.Client, string, string) error                             = indexer.Refresh
	_ func(context.Context, time.Duration, func(context.Context) error, func() error) error = indexer.Serve
)

func TestIndexesAreIndependent(t *testing.T) {
	left := indexer.Open(t.TempDir())
	right := indexer.Open(t.TempDir())
	if err := left.Put("alpha", "left"); err != nil {
		t.Fatal(err)
	}
	if err := right.Put("alpha", "right"); err != nil {
		t.Fatal(err)
	}
	assertText(t, left, "alpha", "left")
	assertText(t, right, "alpha", "right")
	if err := left.Put("alpha", "replaced left"); err != nil {
		t.Fatal(err)
	}
	assertText(t, left, "alpha", "replaced left")
	assertText(t, right, "alpha", "right")
}

func TestPutKeyContract(t *testing.T) {
	for _, key := range []string{"", "Alpha", "a1", "a-b", " a", "a/escape", "é"} {
		t.Run(key, func(t *testing.T) {
			t.Parallel()
			i := indexer.Open(t.TempDir())
			if err := i.Put(key, "text"); !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("Put(%q) = %v; want ErrInvalidRecord", key, err)
			}
			if _, err := i.Get(key); !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("Get(%q) = %v; want ErrInvalidRecord", key, err)
			}
		})
	}
}

func TestApplyCSVOrderedReplacementAndText(t *testing.T) {
	i := indexer.Open(t.TempDir())
	input := "key,text\nalpha,first\nbeta,\"  comma, and\nnewline  \"\nalpha,last\n"
	if err := i.ApplyCSV(strings.NewReader(input)); err != nil {
		t.Fatal(err)
	}
	assertText(t, i, "key", "text") // No header is skipped.
	assertText(t, i, "alpha", "last")
	assertText(t, i, "beta", "  comma, and\nnewline  ")
	if err := i.ApplyCSV(strings.NewReader("")); err != nil {
		t.Fatalf("empty input: %v", err)
	}
	assertText(t, i, "alpha", "last")
}

func TestApplyCSVRejectedReplacementRetainsPrefix(t *testing.T) {
	for _, tc := range []struct {
		name, row string
		invalid   bool
	}{
		{"empty text", "alpha,\n", true},
		{"unicode whitespace", "alpha,\" \t\u2003\"\n", true},
		{"invalid key", "../alpha,rejected\n", true},
		{"one field", "alpha\n", true},
		{"three fields", "alpha,rejected,extra\n", true},
		{"malformed quote", "alpha,\"unterminated\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			i := indexer.Open(t.TempDir())
			if err := i.Put("alpha", "before import"); err != nil {
				t.Fatal(err)
			}
			err := i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,accepted beta\n" + tc.row + "gamma,later\n"))
			if err == nil {
				t.Fatal("rejected row succeeded")
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("error = %v; want ErrInvalidRecord", err)
			}
			assertText(t, i, "alpha", "accepted")
			assertText(t, i, "beta", "accepted beta")
			assertAbsent(t, i, "gamma")
		})
	}
}

func TestApplyCSVReadFailureRetainsPrefix(t *testing.T) {
	wantErr := errors.New("reader failed")
	i := indexer.Open(t.TempDir())
	r := io.MultiReader(strings.NewReader("alpha,accepted\n"), failingReader{wantErr})
	if err := i.ApplyCSV(r); !errors.Is(err, wantErr) {
		t.Fatalf("ApplyCSV = %v; want read cause", err)
	}
	assertText(t, i, "alpha", "accepted")
}

func TestApplyCSVPublicationFailureRetainsPrefix(t *testing.T) {
	root := t.TempDir()
	i := indexer.Open(root)
	if err := os.Mkdir(filepath.Join(root, "beta"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,rejected\ngamma,later\n")); err == nil {
		t.Fatal("replacement of directory succeeded")
	}
	assertText(t, i, "alpha", "accepted")
	assertAbsent(t, i, "gamma")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("directory entries = %v; want alpha and beta only", entries)
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func assertText(t *testing.T, i *indexer.Index, key, want string) {
	t.Helper()
	got, err := i.Get(key)
	if err != nil || got != want {
		t.Fatalf("Get(%q) = %q, %v; want %q", key, got, err, want)
	}
}

func assertAbsent(t *testing.T, i *indexer.Index, key string) {
	t.Helper()
	if got, err := i.Get(key); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Get(%q) = %q, %v; want absent", key, got, err)
	}
}
