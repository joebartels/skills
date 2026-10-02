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

// Compile representative consumers against the exact supported signatures.
var (
	_ func(string) *indexer.Index                  = indexer.Open
	_ func(*indexer.Index, string, string) error   = (*indexer.Index).Put
	_ func(*indexer.Index, string) (string, error) = (*indexer.Index).Get
	_ func(*indexer.Index, io.Reader) error        = (*indexer.Index).ApplyCSV
)

func TestExistingPutAndReplace(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	for _, text := range []string{"first", "x", "", " \t\n"} {
		if err := i.Put("alpha", text); err != nil {
			t.Fatal(err)
		}
		got, err := i.Get("alpha")
		if err != nil || got != text {
			t.Fatalf("Get = %q, %v; want %q", got, err, text)
		}
	}
	for _, key := range []string{"", "../escape", "Alpha", "a1", "é"} {
		if err := i.Put(key, "bad"); !errors.Is(err, indexer.ErrInvalidRecord) {
			t.Errorf("Put(%q) = %v", key, err)
		}
		if _, err := i.Get(key); !errors.Is(err, indexer.ErrInvalidRecord) {
			t.Errorf("Get(%q) = %v", key, err)
		}
	}
}

func TestIndependentIndexRoots(t *testing.T) {
	t.Parallel()
	first, second := indexer.Open(t.TempDir()), indexer.Open(t.TempDir())
	for n, i := range []*indexer.Index{first, second} {
		text := []string{"one", "two"}[n]
		if err := i.Put("alpha", text); err != nil {
			t.Fatal(err)
		}
	}
	assertStored(t, first, "alpha", "one")
	assertStored(t, second, "alpha", "two")
}

func TestApplyCSVSuccess(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	input := "alpha,first\nbeta,\" comma, and\nnewline \"\nalpha, replacement \n"
	if err := i.ApplyCSV(strings.NewReader(input)); err != nil {
		t.Fatal(err)
	}
	assertStored(t, i, "alpha", " replacement ")
	assertStored(t, i, "beta", " comma, and\nnewline ")
	if err := i.ApplyCSV(strings.NewReader("")); err != nil {
		t.Fatalf("empty CSV = %v", err)
	}
}

func TestApplyCSVRejectedPrefix(t *testing.T) {
	cases := []struct {
		name    string
		row     string
		invalid bool
	}{
		{"blank_replacement", "alpha, \t\n", true},
		{"unicode_blank", "alpha,\u2003\n", true},
		{"invalid_key", "Alpha,bad\n", true},
		{"one_field", "alpha\n", true},
		{"three_fields", "alpha,bad,extra\n", true},
		{"malformed_quote", "alpha,\"unterminated\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			i := indexer.Open(t.TempDir())
			if err := i.Put("alpha", "original"); err != nil {
				t.Fatal(err)
			}
			err := i.ApplyCSV(strings.NewReader("alpha,accepted\n" + tc.row + "gamma,later\n"))
			if err == nil || errors.Is(err, indexer.ErrInvalidRecord) != tc.invalid {
				t.Fatalf("ApplyCSV = %v; validation identity want %v", err, tc.invalid)
			}
			assertStored(t, i, "alpha", "accepted")
			if _, err := i.Get("gamma"); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("later row Get = %v; want not exist", err)
			}
		})
	}
}

func TestApplyCSVReadAndWriteFailure(t *testing.T) {
	t.Run("reader", func(t *testing.T) {
		t.Parallel()
		i := indexer.Open(t.TempDir())
		wantErr := errors.New("reader failed")
		r := io.MultiReader(strings.NewReader("alpha,accepted\n"), errorReader{wantErr})
		if err := i.ApplyCSV(r); !errors.Is(err, wantErr) {
			t.Fatalf("ApplyCSV = %v; want reader error", err)
		}
		assertStored(t, i, "alpha", "accepted")
	})
	t.Run("publication", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		blocked := filepath.Join(root, "beta")
		if err := os.Mkdir(blocked, 0700); err != nil {
			t.Fatal(err)
		}
		marker := filepath.Join(blocked, "prior")
		if err := os.WriteFile(marker, []byte("keep me"), 0600); err != nil {
			t.Fatal(err)
		}
		i := indexer.Open(root)
		err := i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,rejected\ngamma,later\n"))
		if err == nil || errors.Is(err, indexer.ErrInvalidRecord) {
			t.Fatalf("write failure = %v", err)
		}
		assertStored(t, i, "alpha", "accepted")
		assertFile(t, marker, "keep me")
		if _, err := i.Get("gamma"); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("later row = %v", err)
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
	})
}

func TestPutWriteFailurePreservesPriorFile(t *testing.T) {
	root := t.TempDir()
	i := indexer.Open(root)
	if err := i.Put("alpha", "old bytes\n"); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(root, 0700) })
	err := i.Put("alpha", "replacement")
	if err == nil {
		t.Skip("filesystem privileges bypass directory write permissions")
	}
	assertStored(t, i, "alpha", "old bytes\n")
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

func assertStored(t *testing.T, i *indexer.Index, key, want string) {
	t.Helper()
	got, err := i.Get(key)
	if err != nil || got != want {
		t.Fatalf("Get(%q) = %q, %v; want %q", key, got, err, want)
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || string(got) != want {
		t.Fatalf("ReadFile(%q) = %q, %v; want %q", path, got, err, want)
	}
}
