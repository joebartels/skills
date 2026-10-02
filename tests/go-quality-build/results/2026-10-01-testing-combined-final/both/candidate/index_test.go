package indexer_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"example.com/indexer"
)

// Representative external consumers retain the exact exported function types.
var (
	_ func(string) *indexer.Index                                                           = indexer.Open
	_ func(*indexer.Index, string, string) error                                            = (*indexer.Index).Put
	_ func(*indexer.Index, string) (string, error)                                          = (*indexer.Index).Get
	_ func(*indexer.Index, io.Reader) error                                                 = (*indexer.Index).ApplyCSV
	_ func(context.Context, time.Duration, func(context.Context) error, func() error) error = indexer.Serve
)

func TestExistingPutAndReplace(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	for _, text := range []string{"first", "x", "", " \t\n任意 text"} {
		if err := i.Put("alpha", text); err != nil {
			t.Fatal(err)
		}
		got, err := i.Get("alpha")
		if err != nil || got != text {
			t.Fatalf("Get = %q, %v; want %q", got, err, text)
		}
	}
	for _, key := range []string{"../escape", "", "Alpha", "a1", "é", "a-b"} {
		if err := i.Put(key, "bad"); !errors.Is(err, indexer.ErrInvalidRecord) {
			t.Errorf("Put(%q) = %v", key, err)
		}
		if _, err := i.Get(key); !errors.Is(err, indexer.ErrInvalidRecord) {
			t.Errorf("Get(%q) = %v", key, err)
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
	assertValue(t, first, "alpha", "first")
	assertValue(t, second, "alpha", "second")
}

func TestApplyCSVSuccess(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	input := "alpha,first\nbeta,\"  comma, and\nnewline  \"\nalpha,last\n"
	if err := i.ApplyCSV(strings.NewReader(input)); err != nil {
		t.Fatal(err)
	}
	assertValue(t, i, "alpha", "last")
	assertValue(t, i, "beta", "  comma, and\nnewline  ")
	if err := i.ApplyCSV(strings.NewReader("")); err != nil {
		t.Fatalf("empty CSV: %v", err)
	}
	assertValue(t, i, "alpha", "last")
}

func TestApplyCSVRejectsAndRetainsPrefix(t *testing.T) {
	cases := []struct {
		name, rejected string
		invalid        bool
	}{
		{"invalid duplicate key", "Alpha,replacement\n", true},
		{"blank duplicate text", "alpha, \t\u2003\n", true},
		{"empty duplicate text", "alpha,\n", true},
		{"missing field", "alpha\n", false},
		{"extra field", "alpha,replacement,extra\n", false},
		{"malformed quote", "alpha,\"unterminated\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			i := indexer.Open(t.TempDir())
			if err := i.Put("alpha", "old"); err != nil {
				t.Fatal(err)
			}
			err := i.ApplyCSV(strings.NewReader("beta,accepted\nalpha,last accepted\n" + tc.rejected + "gamma,later\n"))
			if err == nil {
				t.Fatal("rejected row succeeded")
			}
			if tc.invalid && !errors.Is(err, indexer.ErrInvalidRecord) {
				t.Errorf("error = %v; want ErrInvalidRecord", err)
			}
			assertValue(t, i, "alpha", "last accepted")
			assertValue(t, i, "beta", "accepted")
			if _, err := i.Get("gamma"); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("later row exists: %v", err)
			}
		})
	}
}

type failingReader struct {
	prefix *strings.Reader
	err    error
}

func (r *failingReader) Read(p []byte) (int, error) {
	n, err := r.prefix.Read(p)
	if err == io.EOF {
		return n, r.err
	}
	return n, err
}

func TestApplyCSVReadFailureRetainsPrefix(t *testing.T) {
	t.Parallel()
	i := indexer.Open(t.TempDir())
	want := errors.New("source failed")
	err := i.ApplyCSV(&failingReader{strings.NewReader("alpha,accepted\n"), want})
	if !errors.Is(err, want) {
		t.Fatalf("ApplyCSV = %v; want source cause", err)
	}
	assertValue(t, i, "alpha", "accepted")
}

func TestApplyCSVPublicationFailureStops(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	i := indexer.Open(root)
	if err := os.Mkdir(filepath.Join(root, "beta"), 0700); err != nil {
		t.Fatal(err)
	}
	err := i.ApplyCSV(strings.NewReader("alpha,accepted\nbeta,rejected\ngamma,later\n"))
	if err == nil {
		t.Fatal("writing over directory succeeded")
	}
	assertValue(t, i, "alpha", "accepted")
	if _, err := i.Get("gamma"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("later row exists: %v", err)
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

func assertValue(t *testing.T, i *indexer.Index, key, want string) {
	t.Helper()
	got, err := i.Get(key)
	if err != nil || got != want {
		t.Fatalf("Get(%q) = %q, %v; want %q", key, got, err, want)
	}
}
