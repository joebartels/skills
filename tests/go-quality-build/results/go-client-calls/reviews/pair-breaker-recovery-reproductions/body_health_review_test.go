package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type bodyHealthTransport func(*http.Request) (*http.Response, error)
func (f bodyHealthTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
type bodyHealthReader struct{ err error }
func (r bodyHealthReader) Read([]byte) (int, error) { return 0, r.err }

func TestPairBodyReadErrorIsNeutral(t *testing.T) {
	for _, recovery := range []bool{false, true} {
		t.Run(map[bool]string{false: "consecutive history", true: "recovery"}[recovery], func(t *testing.T) {
			calls := 0
			cause := errors.New("body read failed")
			c := New(&http.Client{Transport: bodyHealthTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				resp := &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("down"))}
				if r.URL.Path == "/read-error" {
					resp.StatusCode = 200
					resp.Body = io.NopCloser(bodyHealthReader{err: cause})
				}
				return resp, nil
			})}, true)
			failures := 2
			if recovery { failures = 3 }
			for i := 0; i < failures; i++ {
				_, err := c.Fetch(context.Background(), "http://dep.test/fail")
				if err == nil || err.Error() != "HTTP 503" { t.Fatalf("initial failure=%v", err) }
			}
			if recovery { time.Sleep(110*time.Millisecond) }
			_, err := c.Fetch(context.Background(), "http://dep.test/read-error")
			if !errors.Is(err, cause) { t.Fatalf("body caller error=%v", err) }
			_, err = c.Fetch(context.Background(), "http://dep.test/fail")
			if err == nil || err.Error() != "HTTP 503" { t.Fatalf("replacement eligible failure=%v", err) }
			before := calls
			_, err = c.Fetch(context.Background(), "http://dep.test/fail")
			t.Logf("after excluded read failure: result=%v calls=%d->%d", err, before, calls)
			if err == nil || calls != before { t.Fatal("read failure reset history or established recovery") }
		})
	}
}
