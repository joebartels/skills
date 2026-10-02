package indexer_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/indexer"
)

// These assignments protect the original function and method signatures.
var (
	_ func(string) *indexer.Index                  = indexer.Open
	_ func(*indexer.Index, string, string) error   = (*indexer.Index).Put
	_ func(*indexer.Index, string) (string, error) = (*indexer.Index).Get
	_ func(*indexer.Index, io.Reader) error        = (*indexer.Index).ApplyCSV
)

func TestExistingPutAndReplace(t *testing.T) {
	i := indexer.Open(t.TempDir())
	for _, text := range []string{"first", "x", "", " \n\t"} {
		if err := i.Put("alpha", text); err != nil {
			t.Fatal(err)
		}
		assertRecord(t, i, "alpha", text)
	}
	for _, key := range []string{"", "../escape", "A", "alpha1", "é", "a/b"} {
		if err := i.Put(key, "bad"); !errors.Is(err, indexer.ErrInvalidRecord) {
			t.Errorf("Put(%q) = %v, want invalid record", key, err)
		}
		if _, err := i.Get(key); !errors.Is(err, indexer.ErrInvalidRecord) {
			t.Errorf("Get(%q) = %v, want invalid record", key, err)
		}
	}
	assertRecord(t, i, "alpha", " \n\t")
}

func TestIndependentRoots(t *testing.T) {
	first, second := indexer.Open(t.TempDir()), indexer.Open(t.TempDir())
	if err := first.Put("alpha", "first"); err != nil {
		t.Fatal(err)
	}
	if err := second.Put("alpha", "second"); err != nil {
		t.Fatal(err)
	}
	assertRecord(t, first, "alpha", "first")
	assertRecord(t, second, "alpha", "second")
}

func TestApplyCSVOrderedRecords(t *testing.T) {
	i := indexer.Open(t.TempDir())
	input := "alpha,first\nbeta,\"  with, comma\nnext line  \"\nalpha,last\n"
	if err := i.ApplyCSV(strings.NewReader(input)); err != nil {
		t.Fatal(err)
	}
	assertRecord(t, i, "alpha", "last")
	assertRecord(t, i, "beta", "  with, comma\nnext line  ")
	if err := i.ApplyCSV(strings.NewReader("")); err != nil {
		t.Fatal(err)
	}
	assertRecord(t, i, "alpha", "last")
}

func TestApplyCSVRetainsAcceptedPrefix(t *testing.T) {
	for _, tc := range []struct {
		name, bad string
		invalid   bool
	}{
		{"key", "Alpha,rejected\n", true},
		{"blank replacement", "alpha,\" \t \"\n", true},
		{"Unicode whitespace", "alpha,\u2003\n", true},
		{"too few fields", "alpha\n", true},
		{"too many fields", "alpha,bad,extra\n", true},
		{"parse", "alpha,\"unterminated\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := indexer.Open(t.TempDir())
			if err := i.Put("alpha", "original"); err != nil {
				t.Fatal(err)
			}
			err := i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,also accepted\n" + tc.bad + "gamma,later\n"))
			if err == nil {
				t.Fatal("missing rejection")
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("error = %v", err)
			}
			assertRecord(t, i, "alpha", "accepted")
			assertRecord(t, i, "beta", "also accepted")
			assertMissing(t, i, "gamma")
		})
	}
}

func TestApplyCSVReadAndWriteFailures(t *testing.T) {
	t.Run("reader failure", func(t *testing.T) {
		i := indexer.Open(t.TempDir())
		cause := errors.New("reader interrupted")
		err := i.ApplyCSV(io.MultiReader(strings.NewReader("alpha,accepted\n"), failingReader{cause}))
		if !errors.Is(err, cause) {
			t.Fatalf("error = %v, want reader cause", err)
		}
		assertRecord(t, i, "alpha", "accepted")
	})
	t.Run("publication failure", func(t *testing.T) {
		root := t.TempDir()
		i := indexer.Open(root)
		if err := os.Mkdir(filepath.Join(root, "beta"), 0700); err != nil {
			t.Fatal(err)
		}
		err := i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,rejected\ngamma,later\n"))
		if err == nil {
			t.Fatal("missing write rejection")
		}
		assertRecord(t, i, "alpha", "accepted")
		assertMissing(t, i, "gamma")
		entries, err := os.ReadDir(root)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 2 {
			t.Fatalf("entries = %v, want alpha and beta only", entries)
		}
	})
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func assertRecord(t *testing.T, i *indexer.Index, key, want string) {
	t.Helper()
	got, err := i.Get(key)
	if err != nil || got != want {
		t.Fatalf("Get(%q) = %q, %v; want %q", key, got, err, want)
	}
}
func assertMissing(t *testing.T, i *indexer.Index, key string) {
	t.Helper()
	if _, err := i.Get(key); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Get(%q) = %v, want missing", key, err)
	}
}
