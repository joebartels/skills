package stages

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestSupplementaryCanceledEmpty(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err := FetchAll(ctx, &http.Client{}, nil, time.Second, time.Second)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty endpoints must succeed without calls: bodies=%v err=%v", got, err)
	}
}
func TestSupplementaryIndependentCustomCause(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	cause := errors.New("custom stop")
	failure := errors.New("independent transport failure")
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) { cancel(cause); return nil, failure })}
	_, err := FetchAll(ctx, client, []string{"http://example.test/a"}, time.Second, time.Second)
	if !errors.Is(err, failure) || !errors.Is(err, cause) || !errors.Is(err, context.Canceled) {
		t.Fatalf("missing independent/cause/classification: %v", err)
	}
}
