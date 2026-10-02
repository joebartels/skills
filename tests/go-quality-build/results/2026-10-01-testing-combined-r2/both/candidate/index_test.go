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

// These assignments protect function types as well as ordinary call syntax.
var (
	_ func(string) *indexer.Index                  = indexer.Open
	_ func(*indexer.Index, string, string) error   = (*indexer.Index).Put
	_ func(*indexer.Index, string) (string, error) = (*indexer.Index).Get
	_ func(*indexer.Index, io.Reader) error        = (*indexer.Index).ApplyCSV
)

func TestExistingPutAndReplace(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	for _, text := range []string{"first", "x", "", "  ", "\x00arbitrary"} {
		if err := i.Put("alpha", text); err != nil {
			t.Fatal(err)
		}
		assertText(t, i, "alpha", text)
	}
	for _, key := range []string{"../escape", "", "Alpha", "alpha1", "é", "a/b"} {
		if err := i.Put(key, "bad"); !errors.Is(err, indexer.ErrInvalidRecord) {
			t.Fatalf("Put(%q) = %v", key, err)
		}
		if _, err := i.Get(key); !errors.Is(err, indexer.ErrInvalidRecord) {
			t.Fatalf("Get(%q) = %v", key, err)
		}
	}
}

func TestIndependentRoots(t *testing.T) {
	t.Parallel()
	first, second := indexer.Open(t.TempDir()), indexer.Open(t.TempDir())
	if err := first.Put("alpha", "first"); err != nil {
		t.Fatal(err)
	}
	if err := second.Put("alpha", "second"); err != nil {
		t.Fatal(err)
	}
	assertText(t, first, "alpha", "first")
	assertText(t, second, "alpha", "second")
}

func TestApplyCSVSuccess(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	if err := i.ApplyCSV(strings.NewReader("alpha,first\nbeta,\"comma, and\nnewline\"\nalpha,  replacement  \n")); err != nil {
		t.Fatal(err)
	}
	assertText(t, i, "alpha", "  replacement  ")
	assertText(t, i, "beta", "comma, and\nnewline")
	if err := i.ApplyCSV(strings.NewReader("")); err != nil {
		t.Fatalf("empty CSV: %v", err)
	}
	assertText(t, i, "alpha", "  replacement  ")
}

func TestApplyCSVRejectedReplacementRetainsPrefix(t *testing.T) {
	cases := []struct {
		name, rejected string
		invalid        bool
	}{
		{"blank text", "alpha,  ", true},
		{"unicode whitespace", "alpha,\u2003", true},
		{"invalid key", "Alpha,rejected", true},
		{"too few fields", "alpha", false},
		{"too many fields", "alpha,rejected,extra", false},
		{"malformed quotes", "alpha,\"unterminated", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			i := indexer.Open(t.TempDir())
			if err := i.Put("alpha", "original"); err != nil {
				t.Fatal(err)
			}
			err := i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,prefix\n" + tc.rejected + "\ngamma,later\n"))
			if err == nil {
				t.Fatal("expected rejection")
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("error = %v; want ErrInvalidRecord", err)
			}
			assertText(t, i, "alpha", "accepted")
			assertText(t, i, "beta", "prefix")
			assertMissing(t, i, "gamma")
		})
	}
}

func TestApplyCSVWriteFailureRetainsPrefix(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "beta"), 0700); err != nil {
		t.Fatal(err)
	}
	i := indexer.Open(root)
	err := i.ApplyCSV(strings.NewReader("alpha,prefix\nbeta,rejected\ngamma,later\n"))
	if err == nil {
		t.Fatal("expected rename failure")
	}
	assertText(t, i, "alpha", "prefix")
	assertMissing(t, i, "gamma")
	info, err := os.Stat(filepath.Join(root, "beta"))
	if err != nil || !info.IsDir() {
		t.Fatalf("obstructing directory changed: %v, %v", info, err)
	}
	assertNoTemps(t, root)
}

func assertText(t *testing.T, i *indexer.Index, key, want string) {
	t.Helper()
	got, err := i.Get(key)
	if err != nil || got != want {
		t.Fatalf("Get(%q) = %q, %v; want %q", key, got, err, want)
	}
}
func assertMissing(t *testing.T, i *indexer.Index, key string) {
	t.Helper()
	got, err := i.Get(key)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected later key %q = %q, %v", key, got, err)
	}
}
func assertNoTemps(t *testing.T, root string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(root, ".record-*"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("temporary files leaked: %v, %v", matches, err)
	}
}

func TestApplyCSVReadFailureRetainsAcceptedPrefix(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	readErr := errors.New("CSV input failed after prefix")
	input := io.MultiReader(strings.NewReader("alpha,prefix\n"), brokenReader{err: readErr})
	if err := i.ApplyCSV(input); !errors.Is(err, readErr) {
		t.Fatalf("read error = %v; want %v", err, readErr)
	}
	assertText(t, i, "alpha", "prefix")
	assertMissing(t, i, "beta")
}
