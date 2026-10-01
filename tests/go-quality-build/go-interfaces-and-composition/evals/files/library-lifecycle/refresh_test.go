package refreshkit_test

import (
	"context"
	"errors"
	refreshkit "example.com/library-lifecycle"
	"testing"
)

func TestRefresh(t *testing.T) {
	want := errors.New("refresh failed")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	r, err := refreshkit.New(func(got context.Context) error {
		calls++
		if got != ctx {
			t.Error("context was replaced")
		}
		return want
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatal("unexpected early callback")
	}
	if err := r.Refresh(ctx); err != want {
		t.Fatalf("error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d", calls)
	}
}
func TestNilCallback(t *testing.T) {
	if _, err := refreshkit.New(nil); err == nil {
		t.Fatal("nil callback accepted")
	}
}
