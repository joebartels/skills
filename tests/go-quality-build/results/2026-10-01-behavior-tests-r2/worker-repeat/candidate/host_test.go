package sweeper

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSweepFailureReleases(t *testing.T) {
	want := errors.New("sweep failed")
	released := false
	err := Run(context.Background(), time.Second, func(context.Context) error {
		return want
	}, func() error {
		released = true
		return nil
	})
	if !errors.Is(err, want) || !released {
		t.Fatalf("Run = %v, released = %v", err, released)
	}
}
