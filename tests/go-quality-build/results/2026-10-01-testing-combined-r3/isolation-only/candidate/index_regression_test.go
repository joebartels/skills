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

// Function assignments exercise the existing consumer-visible signatures.
var (
	_ func(string) *indexer.Index                  = indexer.Open
	_ func(*indexer.Index, string, string) error   = (*indexer.Index).Put
	_ func(*indexer.Index, string) (string, error) = (*indexer.Index).Get
	_ func(*indexer.Index, io.Reader) error        = (*indexer.Index).ApplyCSV
)

func TestIndependentIndexes(t *testing.T) {
	t.Parallel()
	first := indexer.Open(t.TempDir())
	second := indexer.Open(t.TempDir())
	if err := first.Put("same", "first\x00\n"); err != nil {
		t.Fatal(err)
	}
	if err := second.Put("same", "second"); err != nil {
		t.Fatal(err)
	}
	assertValue(t, first, "same", "first\x00\n")
	assertValue(t, second, "same", "second")
}

func TestPutRetainsPriorFileOnFailure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	index := indexer.Open(root)
	if err := index.Put("alpha", "original"); err != nil {
		t.Fatal(err)
	}
	if err := index.Put("Alpha", "rejected"); !errors.Is(err, indexer.ErrInvalidRecord) {
		t.Fatalf("error = %v", err)
	}
	assertValue(t, index, "alpha", "original")
	// A directory at a valid destination forces rename failure independently
	// of user privileges and platform permission modes.
	blocked := filepath.Join(root, "blocked")
	if err := os.Mkdir(blocked, 0700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(blocked, "marker")
	if err := os.WriteFile(marker, []byte("retained"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := index.Put("blocked", "new"); err == nil {
		t.Fatal("expected publication failure")
	}
	if got, err := os.ReadFile(marker); err != nil || string(got) != "retained" {
		t.Fatalf("marker = %q, %v", got, err)
	}
	assertNoTemps(t, root)
}

func TestApplyCSV(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, input        string
		want               map[string]string
		absent             []string
		wantError, invalid bool
	}{
		{name: "empty", input: "", want: map[string]string{"alpha": "old"}},
		{name: "ordered_duplicates_and_exact_text", input: "alpha,first\nalpha,second\nbeta,\" line one\nline two,quoted \"\n", want: map[string]string{"alpha": "second", "beta": " line one\nline two,quoted "}},
		{name: "invalid_replacement", input: "alpha,accepted\nalpha, \t\nbeta,later\n", want: map[string]string{"alpha": "accepted"}, absent: []string{"beta"}, wantError: true, invalid: true},
		{name: "invalid_key", input: "alpha,accepted\n../escape,bad\nbeta,later\n", want: map[string]string{"alpha": "accepted"}, absent: []string{"beta"}, wantError: true, invalid: true},
		{name: "unicode_blank", input: "alpha,\u2003\n", want: map[string]string{"alpha": "old"}, wantError: true, invalid: true},
		{name: "too_few_fields", input: "alpha,accepted\nbeta\ngamma,later\n", want: map[string]string{"alpha": "accepted"}, absent: []string{"beta", "gamma"}, wantError: true, invalid: true},
		{name: "too_many_fields", input: "alpha,accepted\nbeta,one,two\ngamma,later\n", want: map[string]string{"alpha": "accepted"}, absent: []string{"beta", "gamma"}, wantError: true, invalid: true},
		{name: "malformed_csv", input: "alpha,accepted\nbeta,\"unterminated\ngamma,later\n", want: map[string]string{"alpha": "accepted"}, absent: []string{"beta", "gamma"}, wantError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			index := indexer.Open(t.TempDir())
			if err := index.Put("alpha", "old"); err != nil {
				t.Fatal(err)
			}
			err := index.ApplyCSV(strings.NewReader(tc.input))
			if (err != nil) != tc.wantError {
				t.Fatalf("ApplyCSV error = %v; want error %v", err, tc.wantError)
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Fatalf("error identity = %v", err)
			}
			for key, value := range tc.want {
				assertValue(t, index, key, value)
			}
			for _, key := range tc.absent {
				if _, err := index.Get(key); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("later key %q error = %v", key, err)
				}
			}
		})
	}
}

func TestApplyCSVWriteAndReadFailures(t *testing.T) {
	t.Parallel()
	t.Run("write", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := os.Mkdir(filepath.Join(root, "blocked"), 0700); err != nil {
			t.Fatal(err)
		}
		index := indexer.Open(root)
		if err := index.ApplyCSV(strings.NewReader("alpha,accepted\nblocked,rejected\nbeta,later\n")); err == nil {
			t.Fatal("expected write failure")
		}
		assertValue(t, index, "alpha", "accepted")
		if _, err := index.Get("beta"); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("later write: %v", err)
		}
		assertNoTemps(t, root)
	})
	t.Run("reader", func(t *testing.T) {
		t.Parallel()
		index := indexer.Open(t.TempDir())
		readErr := errors.New("source interrupted")
		err := index.ApplyCSV(io.MultiReader(strings.NewReader("alpha,accepted\n"), failingReader{readErr}))
		if !errors.Is(err, readErr) {
			t.Fatalf("read error = %v", err)
		}
		assertValue(t, index, "alpha", "accepted")
	})
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func assertValue(t *testing.T, index *indexer.Index, key, want string) {
	t.Helper()
	got, err := index.Get(key)
	if err != nil || got != want {
		t.Fatalf("Get(%q) = %q, %v; want %q", key, got, err, want)
	}
}

func assertNoTemps(t *testing.T, root string) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(root, ".record-*"))
	if err != nil || len(files) != 0 {
		t.Fatalf("temporary files = %v, %v", files, err)
	}
}
