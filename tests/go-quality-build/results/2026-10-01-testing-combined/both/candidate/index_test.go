package indexer_test

import (
	"errors"
	"example.com/indexer"
	"testing"
)

func TestExistingPutAndReplace(t *testing.T) {
	i := indexer.Open(t.TempDir())
	for _, text := range []string{"first", "x", "", " \t\n\u2003", "arbitrary\x00text"} {
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
