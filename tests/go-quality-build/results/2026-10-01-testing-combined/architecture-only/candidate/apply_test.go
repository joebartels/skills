package indexer_test

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/indexer"
)

// Compile the exported function and method types as an external consumer.
var (
	_ func(string) *indexer.Index                  = indexer.Open
	_ func(*indexer.Index, string, string) error   = (*indexer.Index).Put
	_ func(*indexer.Index, string) (string, error) = (*indexer.Index).Get
	_ func(*indexer.Index, io.Reader) error        = (*indexer.Index).ApplyCSV
)

func assertText(t *testing.T, i *indexer.Index, key, want string) {
	t.Helper()
	got, err := i.Get(key)
	if err != nil || got != want {
		t.Fatalf("Get(%q) = %q, %v; want %q", key, got, err, want)
	}
}

func assertMissing(t *testing.T, i *indexer.Index, key string) {
	t.Helper()
	if _, err := i.Get(key); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Get(%q) error = %v; want missing file", key, err)
	}
}

func TestIndependentIndexesAndPutText(t *testing.T) {
	a, b := indexer.Open(t.TempDir()), indexer.Open(t.TempDir())
	for _, pair := range []struct {
		index *indexer.Index
		text  string
	}{{a, ""}, {b, "\x00 arbitrary\ntext \t"}} {
		if err := pair.index.Put("alpha", pair.text); err != nil {
			t.Fatal(err)
		}
	}
	assertText(t, a, "alpha", "")
	assertText(t, b, "alpha", "\x00 arbitrary\ntext \t")
	for _, key := range []string{"", "Alpha", "a1", "a/b", "../alpha", "é"} {
		if err := a.Put(key, "changed"); !errors.Is(err, indexer.ErrInvalidRecord) {
			t.Errorf("Put(%q) = %v; want invalid record", key, err)
		}
		if _, err := a.Get(key); !errors.Is(err, indexer.ErrInvalidRecord) {
			t.Errorf("Get(%q) = %v; want invalid record", key, err)
		}
	}
	assertText(t, a, "alpha", "")
}

func TestApplyCSVPreservesTextAndOrder(t *testing.T) {
	i := indexer.Open(t.TempDir())
	input := "alpha,first\nbeta,\"  comma, newline\ntext  \"\nalpha,replacement\n"
	if err := i.ApplyCSV(strings.NewReader(input)); err != nil {
		t.Fatal(err)
	}
	assertText(t, i, "alpha", "replacement")
	assertText(t, i, "beta", "  comma, newline\ntext  ")
	if err := i.ApplyCSV(strings.NewReader("")); err != nil {
		t.Fatalf("empty CSV: %v", err)
	}
	assertText(t, i, "alpha", "replacement")
}

func TestApplyCSVRejectedReplacementRetainsAcceptedPrefix(t *testing.T) {
	cases := []struct {
		name, rejected string
		invalid        bool
		parse          bool
	}{
		{"blank text", "alpha,\" \t\u2003\"\n", true, false},
		{"invalid key", "Alpha,rejected\n", true, false},
		{"one field", "alpha\n", true, true},
		{"three fields", "alpha,rejected,extra\n", true, true},
		{"broken quote", "alpha,\"unterminated\n", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			i := indexer.Open(t.TempDir())
			if err := i.Put("alpha", "before"); err != nil {
				t.Fatal(err)
			}
			err := i.ApplyCSV(strings.NewReader("alpha,accepted\n" + tc.rejected + "gamma,later\n"))
			if err == nil {
				t.Fatal("rejected row succeeded")
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("error = %v; want invalid-record identity", err)
			}
			var parseErr *csv.ParseError
			if tc.parse && !errors.As(err, &parseErr) {
				t.Fatalf("error = %v; want CSV parse error", err)
			}
			assertText(t, i, "alpha", "accepted")
			assertMissing(t, i, "gamma")
		})
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestApplyCSVReadFailureRetainsPrefix(t *testing.T) {
	i := indexer.Open(t.TempDir())
	failure := errors.New("source failed")
	r := io.MultiReader(strings.NewReader("alpha,accepted\n"), failingReader{failure})
	if err := i.ApplyCSV(r); !errors.Is(err, failure) {
		t.Fatalf("ApplyCSV = %v; want source failure", err)
	}
	assertText(t, i, "alpha", "accepted")
}

func TestApplyCSVWriteFailureStopsAtPrefix(t *testing.T) {
	root := t.TempDir()
	blocked := filepath.Join(root, "beta")
	if err := os.Mkdir(blocked, 0700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(blocked, "marker")
	if err := os.WriteFile(marker, []byte("untouched"), 0600); err != nil {
		t.Fatal(err)
	}
	i := indexer.Open(root)
	if err := i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,rejected\ngamma,later\n")); err == nil {
		t.Fatal("writing over a directory succeeded")
	}
	assertText(t, i, "alpha", "accepted")
	assertMissing(t, i, "gamma")
	if got, err := os.ReadFile(marker); err != nil || string(got) != "untouched" {
		t.Fatalf("blocked destination = %q, %v", got, err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".record-") {
			t.Errorf("temporary file leaked: %s", entry.Name())
		}
	}
}
