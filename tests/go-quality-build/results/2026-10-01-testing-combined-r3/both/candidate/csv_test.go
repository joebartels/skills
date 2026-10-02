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

// These assignments exercise the public function types used by consumers.
var (
	_ func(string) *indexer.Index                  = indexer.Open
	_ func(*indexer.Index, string, string) error   = (*indexer.Index).Put
	_ func(*indexer.Index, string) (string, error) = (*indexer.Index).Get
	_ func(*indexer.Index, io.Reader) error        = (*indexer.Index).ApplyCSV
)

func TestIndependentIndexes(t *testing.T) {
	t.Parallel()
	a, b := indexer.Open(t.TempDir()), indexer.Open(t.TempDir())
	if err := a.Put("alpha", "one"); err != nil {
		t.Fatal(err)
	}
	if err := b.Put("alpha", "two"); err != nil {
		t.Fatal(err)
	}
	assertText(t, a, "alpha", "one")
	assertText(t, b, "alpha", "two")
}

func TestPutArbitraryText(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	text := strings.Repeat("large text ", 1024) + "\x00\u2003\n"
	if err := i.Put("alpha", text); err != nil {
		t.Fatal(err)
	}
	assertText(t, i, "alpha", text)
}

func TestApplyCSVSuccess(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	input := "alpha,first\nbeta,\"comma, and\nnewline\"\nalpha, final \n"
	if err := i.ApplyCSV(strings.NewReader(input)); err != nil {
		t.Fatal(err)
	}
	assertText(t, i, "alpha", " final ")
	assertText(t, i, "beta", "comma, and\nnewline")
	if err := i.ApplyCSV(strings.NewReader("")); err != nil {
		t.Fatalf("empty CSV: %v", err)
	}
	assertText(t, i, "alpha", " final ")
}

func TestApplyCSVRejectedPrefix(t *testing.T) {
	for _, tc := range []struct {
		name, row string
		invalid   bool
	}{
		{"key", "Alpha,rejected\n", true},
		{"blank replacement", "alpha, \t\u2003\n", true},
		{"missing field", "alpha\n", true},
		{"extra field", "alpha,rejected,extra\n", true},
		{"malformed quote", "alpha,\"unterminated\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			i := indexer.Open(t.TempDir())
			if err := i.Put("alpha", "seed"); err != nil {
				t.Fatal(err)
			}
			err := i.ApplyCSV(strings.NewReader("alpha,accepted\n" + tc.row + "later,forbidden\n"))
			if err == nil {
				t.Fatal("accepted invalid CSV")
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("error = %v; want ErrInvalidRecord", err)
			}
			assertText(t, i, "alpha", "accepted")
			assertMissing(t, i, "later")
		})
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestApplyCSVReadFailureKeepsPrefix(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	wantErr := errors.New("source read failed")
	err := i.ApplyCSV(io.MultiReader(strings.NewReader("alpha,accepted\n"), failingReader{wantErr}))
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v; want source failure", err)
	}
	assertText(t, i, "alpha", "accepted")
}

func TestApplyCSVPublicationFailureKeepsPrefix(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "beta"), 0700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(root, "beta", "marker")
	if err := os.WriteFile(marker, []byte("untouched"), 0600); err != nil {
		t.Fatal(err)
	}
	i := indexer.Open(root)
	err := i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,rejected\nlater,forbidden\n"))
	if err == nil {
		t.Fatal("accepted replacement over directory")
	}
	assertText(t, i, "alpha", "accepted")
	assertMissing(t, i, "later")
	got, err := os.ReadFile(marker)
	if err != nil || string(got) != "untouched" {
		t.Fatalf("marker = %q, %v", got, err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".record-") {
			t.Errorf("temporary file remains: %s", entry.Name())
		}
	}
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
	if got, err := i.Get(key); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Get(%q) = %q, %v; want absent", key, got, err)
	}
}
