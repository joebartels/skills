package client

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"
)

// The HTTP client supports pre-context custom RoundTrippers through CancelRequest.
// The 700ms release keeps this reproduction bounded even if that facility is lost.
type reviewLegacyTransport struct {
	stopped chan struct{}
	once    sync.Once
	calls   int
}

func (r *reviewLegacyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.calls++
	defer req.Body.Close()
	<-r.stopped
	return nil, context.DeadlineExceeded
}
func (r *reviewLegacyTransport) CancelRequest(*http.Request) { r.once.Do(func() { close(r.stopped) }) }
func TestReviewSubmitLegacyClientTimeout(t *testing.T) {
	base := &reviewLegacyTransport{stopped: make(chan struct{})}
	client := &http.Client{Timeout: 25 * time.Millisecond, Transport: base}
	directReq, _ := http.NewRequest("POST", "http://example.invalid", http.NoBody)
	start := time.Now()
	_, directErr := client.Do(directReq)
	t.Logf("direct client: elapsed=%v error=%v", time.Since(start), directErr)
	if time.Since(start) > 200*time.Millisecond {
		t.Fatal("control client timeout failed")
	}
	base = &reviewLegacyTransport{stopped: make(chan struct{})}
	client.Transport = base
	timer := time.AfterFunc(700*time.Millisecond, func() { base.CancelRequest(nil) })
	defer timer.Stop()
	start = time.Now()
	_, err := Submit(context.Background(), client, "http://example.invalid", "", nil, false)
	elapsed := time.Since(start)
	t.Logf("Submit: elapsed=%v calls=%d deadline=%t timeout=%v error=%v", elapsed, base.calls, errors.Is(err, context.DeadlineExceeded), client.Timeout, err)
	if elapsed > 200*time.Millisecond {
		t.Error("supplied 25ms client timeout lost; total 500ms budget also exceeded")
	}
}
