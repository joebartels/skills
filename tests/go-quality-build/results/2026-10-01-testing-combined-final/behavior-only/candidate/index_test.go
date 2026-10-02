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

// Function assignments also protect consumer signatures, not only call syntax.
var (
	_ func(string) *indexer.Index                  = indexer.Open
	_ func(*indexer.Index, string, string) error   = (*indexer.Index).Put
	_ func(*indexer.Index, string) (string, error) = (*indexer.Index).Get
	_ func(*indexer.Index, io.Reader) error        = (*indexer.Index).ApplyCSV
)

func TestExistingPutAndReplace(t *testing.T) {
	i := indexer.Open(t.TempDir())
	for _, text := range []string{"first", "x", ""} {
		if err := i.Put("alpha", text); err != nil {
			t.Fatal(err)
		}
		got, err := i.Get("alpha")
		if err != nil || got != text {
			t.Fatalf("Get = %q, %v; want %q", got, err, text)
		}
	}
	if err := i.Put("../escape", "bad"); !errors.Is(err, indexer.ErrInvalidRecord) {
		t.Fatalf("invalid key = %v", err)
	}
}

func TestPutKeyRulesAndIndependentRoots(t *testing.T) {
	for _, key := range []string{"", "Alpha", "alpha1", "a-b", "é", "../alpha", "a/b"} {
		t.Run(key, func(t *testing.T) {
			i := indexer.Open(t.TempDir())
			if err := i.Put(key, "text"); !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("Put(%q) = %v", key, err)
			}
			if _, err := i.Get(key); !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("Get(%q) = %v", key, err)
			}
		})
	}
	a, b := indexer.Open(t.TempDir()), indexer.Open(t.TempDir())
	for _, tc := range []struct {
		index *indexer.Index
		text  string
	}{{a, "one\n"}, {b, "two\x00"}} {
		if err := tc.index.Put("alpha", tc.text); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		index *indexer.Index
		text  string
	}{{a, "one\n"}, {b, "two\x00"}} {
		assertText(t, tc.index, "alpha", tc.text)
	}
}

func TestApplyCSVSuccess(t *testing.T) {
	i := indexer.Open(t.TempDir())
	if err := i.ApplyCSV(strings.NewReader("alpha,first\nbeta,\" line one,\nline two \"\nalpha,last\n")); err != nil {
		t.Fatal(err)
	}
	assertText(t, i, "alpha", "last")
	assertText(t, i, "beta", " line one,\nline two ")
	if err := i.ApplyCSV(strings.NewReader("")); err != nil {
		t.Fatal(err)
	}
	assertText(t, i, "alpha", "last")
}

func TestApplyCSVRejectsReplacementAndRetainsPrefix(t *testing.T) {
	for _, tc := range []struct {
		name, row string
		invalid   bool
	}{
		{"blank", "alpha,\" \t\"\n", true},
		{"invalid key", "Alpha,replacement\n", true},
		{"missing field", "alpha\n", false},
		{"extra field", "alpha,replacement,extra\n", false},
		{"malformed quote", "alpha,\"unterminated\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := indexer.Open(t.TempDir())
			if err := i.Put("alpha", "original"); err != nil {
				t.Fatal(err)
			}
			err := i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,prefix\n" + tc.row + "gamma,later\n"))
			if err == nil {
				t.Fatal("expected rejection")
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("identity lost: %v", err)
			}
			assertText(t, i, "alpha", "accepted")
			assertText(t, i, "beta", "prefix")
			if _, err := i.Get("gamma"); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("later row exists: %v", err)
			}
		})
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestApplyCSVReadFailure(t *testing.T) {
	i := indexer.Open(t.TempDir())
	want := errors.New("input interrupted")
	err := i.ApplyCSV(io.MultiReader(strings.NewReader("alpha,accepted\n"), failingReader{want}))
	if !errors.Is(err, want) {
		t.Fatalf("ApplyCSV = %v, want input cause", err)
	}
	assertText(t, i, "alpha", "accepted")
}

func TestApplyCSVWriteFailureStopsLaterRows(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "beta"), 0700); err != nil {
		t.Fatal(err)
	}
	i := indexer.Open(root)
	if err := i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,rejected\ngamma,later\n")); err == nil {
		t.Fatal("expected rename failure")
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
		t.Fatalf("unexpected temporary files: %v", entries)
	}
}

func assertText(t *testing.T, i *indexer.Index, key, want string) {
	t.Helper()
	got, err := i.Get(key)
	if err != nil || got != want {
		t.Fatalf("Get(%q) = %q, %v; want %q", key, got, err, want)
	}
}
