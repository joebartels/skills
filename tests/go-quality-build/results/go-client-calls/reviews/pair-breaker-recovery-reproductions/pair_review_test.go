package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

type pairReviewTransport func(*http.Request) (*http.Response, error)

func (f pairReviewTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func pairReviewReply(code int) *http.Response {
	return &http.Response{StatusCode: code, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("ok"))}
}

func TestPairSuppliedRedirectPolicy(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[enabled], func(t *testing.T) {
			calls, redirects := 0, 0
			h := &http.Client{
				Transport: pairReviewTransport(func(*http.Request) (*http.Response, error) {
					calls++
					r := pairReviewReply(302)
					r.Header.Set("Location", "http://dep.test/end")
					return r, nil
				}),
				CheckRedirect: func(*http.Request, []*http.Request) error {
					redirects++
					return http.ErrUseLastResponse
				},
			}
			_, err := New(h, enabled).Fetch(context.Background(), "http://dep.test/start")
			t.Logf("result=%v calls=%d supplied redirect callbacks=%d", err, calls, redirects)
			if err == nil || err.Error() != "HTTP 302" || calls != 1 || redirects != 1 {
				t.Fatal("Fetch bypassed the supplied redirect policy")
			}
		})
	}
}

func TestPairOperationDeadline503IsNeutral(t *testing.T) {
	var calls atomic.Int32
	c := New(&http.Client{Transport: pairReviewTransport(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		if r.URL.Path == "/deadline" {
			<-r.Context().Done()
		}
		return pairReviewReply(503), nil
	})}, true)
	for i := 0; i < 2; i++ {
		_, err := c.Fetch(context.Background(), "http://dep.test/fail")
		if err == nil || err.Error() != "HTTP 503" { t.Fatalf("eligible failure=%v", err) }
	}
	_, err := c.Fetch(context.Background(), "http://dep.test/deadline")
	if err == nil || err.Error() != "HTTP 503" { t.Fatalf("deadline caller result=%v", err) }
	before := calls.Load()
	_, err = c.Fetch(context.Background(), "http://dep.test/fail")
	t.Logf("after operation deadline: result=%v calls=%d->%d", err, before, calls.Load())
	if err == nil || err.Error() != "HTTP 503" || calls.Load() != before+1 {
		t.Fatal("operation deadline counted as eligible dependency failure")
	}
	before = calls.Load()
	_, err = c.Fetch(context.Background(), "http://dep.test/fail")
	if err == nil || calls.Load() != before { t.Fatal("third eligible failure did not open") }
}

func TestPairOversized200Observation(t *testing.T) {
	h := &http.Client{Transport: pairReviewTransport(func(*http.Request) (*http.Response, error) {
		r := pairReviewReply(200)
		r.Body = io.NopCloser(strings.NewReader(strings.Repeat("x", 4097)))
		return r, nil
	})}
	data, err := New(h, false).Fetch(context.Background(), "http://dep.test/data")
	t.Logf("oversized 200: returned bytes=%d error=%v", len(data), err)
	if len(data) > 4096 { t.Fatal("result bound exceeded") }
}
