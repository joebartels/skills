package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var _ func(context.Context, *http.Client, string) ([]byte, error) = Fetch

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackedBody struct {
	reader   io.Reader
	read     int
	closes   int
	closeErr error
	onClose  func()
}

func (b *trackedBody) Read(p []byte) (int, error) {
	n, err := b.reader.Read(p)
	b.read += n
	return n, err
}

func (b *trackedBody) Close() error {
	b.closes++
	if b.onClose != nil {
		b.onClose()
	}
	return b.closeErr
}

func response(status int, body io.ReadCloser, hint string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Retry-After": []string{hint}}, Body: body}
}

func TestFetchStatusPolicy(t *testing.T) {
	for _, status := range []int{200, 204, 206, 400, 429, 500, 502, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			var bodies []*trackedBody
			c := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				b := &trackedBody{reader: strings.NewReader("result")}
				bodies = append(bodies, b)
				return response(status, b, "0"), nil
			})}
			got, err := Fetch(context.Background(), c, "http://example.test")
			wantCalls := 1
			if status == 429 || status == 503 {
				wantCalls = 3
			}
			if calls != wantCalls {
				t.Errorf("application attempts = %d, want %d", calls, wantCalls)
			}
			if status == 200 {
				if err != nil || string(got) != "result" {
					t.Errorf("Fetch = %q, %v", got, err)
				}
			} else if got != nil || err == nil || !strings.Contains(err.Error(), fmt.Sprintf("HTTP %d", status)) {
				t.Errorf("Fetch = %q, %v, want HTTP %d rejection", got, err, status)
			}
			for i, b := range bodies {
				if b.closes != 1 {
					t.Errorf("body %d closes = %d, want 1", i, b.closes)
				}
			}
		})
	}
}

func TestFetchRetriesShareContextAndCloseBodies(t *testing.T) {
	type key struct{}
	parent := context.WithValue(context.Background(), key{}, "value")
	calls := 0
	var deadline time.Time
	var previous *trackedBody
	var boundaryErr error
	c := &http.Client{Timeout: time.Second}
	c.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		d, ok := r.Context().Deadline()
		if !ok || r.Method != http.MethodGet || r.Context().Value(key{}) != "value" {
			boundaryErr = errors.New("request lost method, value, or deadline")
		}
		if calls == 1 {
			deadline = d
		} else if !d.Equal(deadline) || previous.closes != 1 {
			boundaryErr = errors.New("retry reset deadline or kept previous body open")
		}
		previous = &trackedBody{reader: strings.NewReader("ok")}
		return response([]int{429, 503, 200}[calls-1], previous, "0"), nil
	})
	start := time.Now()
	got, err := Fetch(parent, c, "http://example.test")
	if err != nil || string(got) != "ok" || calls != 3 || boundaryErr != nil {
		t.Fatalf("Fetch = %q, %v; attempts %d; boundary %v", got, err, calls, boundaryErr)
	}
	if d := deadline.Sub(start); d < 200*time.Millisecond || d > 300*time.Millisecond {
		t.Errorf("operation budget = %v, want 250ms", d)
	}
	if c.Timeout != time.Second || previous.closes != 1 {
		t.Errorf("client timeout = %v, final closes = %d", c.Timeout, previous.closes)
	}
}

func TestFetchRetryAfter(t *testing.T) {
	for _, tc := range []struct {
		name, hint string
		unfit      bool
		fallback   bool
	}{
		{"zero", "0", false, false},
		{"past date", "Mon, 02 Jan 2006 15:04:05 GMT", false, false},
		{"seconds cannot fit", "1", true, false},
		{"duration overflow", "9223372037", true, false},
		{"integer overflow", "999999999999999999999999999999", true, false},
		{"future date", time.Now().Add(time.Hour).UTC().Format(http.TimeFormat), true, false},
		{"missing", "", false, true},
		{"malformed", "1.5", false, true},
		{"negative", "-1", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			var first, second time.Time
			c := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				status := 200
				if calls == 1 {
					first = time.Now()
					status = 503
				} else {
					second = time.Now()
				}
				return response(status, io.NopCloser(strings.NewReader("ok")), tc.hint), nil
			})}
			got, err := Fetch(context.Background(), c, "http://example.test")
			if tc.unfit {
				if calls != 1 || got != nil || err == nil || !strings.Contains(err.Error(), "HTTP 503") {
					t.Fatalf("unfit hint: Fetch = %q, %v; attempts %d", got, err, calls)
				}
				if elapsed := time.Since(first); elapsed > 150*time.Millisecond {
					t.Errorf("unfit hint waited %v", elapsed)
				}
				return
			}
			if err != nil || string(got) != "ok" || calls != 2 {
				t.Fatalf("Fetch = %q, %v; attempts %d", got, err, calls)
			}
			if tc.fallback && second.Sub(first) < 10*time.Millisecond {
				t.Errorf("fallback retried after %v, want at least 10ms", second.Sub(first))
			}
		})
	}
}

func TestFetchWaitsForFutureHTTPDate(t *testing.T) {
	// HTTP dates have second precision; arrange a future date within the 250ms budget.
	retryAt := time.Now().UTC().Truncate(time.Second).Add(time.Second)
	if wait := time.Until(retryAt) - 150*time.Millisecond; wait > 0 {
		timer := time.NewTimer(wait)
		<-timer.C
	}
	if !time.Now().Before(retryAt) {
		t.Skip("scheduler passed the future HTTP-date window")
	}
	calls := 0
	var retriedAt time.Time
	c := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		status := 503
		if calls == 2 {
			status = 200
			retriedAt = time.Now()
		}
		return response(status, io.NopCloser(strings.NewReader("ok")), retryAt.Format(http.TimeFormat)), nil
	})}
	got, err := Fetch(context.Background(), c, "http://example.test")
	if err != nil || string(got) != "ok" || calls != 2 {
		t.Fatalf("Fetch = %q, %v; attempts %d", got, err, calls)
	}
	if retriedAt.Before(retryAt) {
		t.Errorf("retry at %v preceded HTTP-date %v", retriedAt, retryAt)
	}
}

func TestFetchCancellationRetainsRejectionAndCause(t *testing.T) {
	cause := errors.New("caller stopped")
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	calls := 0
	body := &trackedBody{reader: strings.NewReader("last rejection"), onClose: func() { cancel(cause) }}
	c := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return response(503, body, ""), nil
	})}
	_, err := Fetch(ctx, c, "http://example.test")
	if err == nil || !strings.Contains(err.Error(), "HTTP 503") || !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
		t.Fatalf("Fetch error = %v, want rejection, cancellation and cause", err)
	}
	if calls != 1 || body.closes != 1 {
		t.Errorf("attempts = %d, closes = %d", calls, body.closes)
	}
}

func TestFetchCancellationDuringFallback(t *testing.T) {
	cause := errors.New("stop during retry wait")
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	done := make(chan struct{})
	var timer *time.Timer
	defer func() {
		if timer != nil && !timer.Stop() {
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Error("cancellation callback did not finish")
			}
		}
	}()
	calls := 0
	body := &trackedBody{reader: strings.NewReader("busy"), onClose: func() {
		// Start cancellation after body ownership ends, during the 10ms fallback.
		timer = time.AfterFunc(time.Millisecond, func() {
			cancel(cause)
			close(done)
		})
	}}
	c := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls != 1 {
			return nil, errors.New("unexpected retry")
		}
		return response(429, body, "malformed"), nil
	})}
	_, err := Fetch(ctx, c, "http://example.test")
	if calls != 1 || err == nil || !strings.Contains(err.Error(), "HTTP 429") || !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
		t.Fatalf("attempts = %d, error = %v, want retained rejection and wait cancellation", calls, err)
	}
}

type readerFunc func([]byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) { return f(p) }

func TestFetchBodyWorkAndRetryShareBudget(t *testing.T) {
	calls := 0
	var deadline time.Time
	var bodies []*trackedBody
	var boundaryErr error
	c := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		d, _ := r.Context().Deadline()
		if calls == 1 {
			deadline = d
		} else if !d.Equal(deadline) {
			boundaryErr = errors.New("body/retry work reset the deadline")
		}
		attempt := calls
		body := &trackedBody{reader: readerFunc(func(p []byte) (int, error) {
			if attempt == 1 {
				timer := time.NewTimer(100 * time.Millisecond)
				defer timer.Stop()
				select {
				case <-timer.C:
					return 0, io.EOF
				case <-r.Context().Done():
					return 0, r.Context().Err()
				}
			}
			<-r.Context().Done()
			return 0, r.Context().Err()
		})}
		bodies = append(bodies, body)
		return response(503, body, ""), nil
	})}
	_, err := Fetch(context.Background(), c, "http://example.test")
	if calls != 2 || boundaryErr != nil || err == nil || !strings.Contains(err.Error(), "HTTP 503") || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("attempts = %d, boundary = %v, error = %v", calls, boundaryErr, err)
	}
	for i, body := range bodies {
		if body.closes != 1 {
			t.Errorf("body %d closes = %d", i, body.closes)
		}
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

func TestFetchBoundedBodiesAndErrors(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		size     int
		readErr  error
		closeErr error
	}{
		{name: "success at limit", status: 200, size: 4096},
		{name: "oversized success", status: 200, size: 10000},
		{name: "bounded rejection", status: 400, size: 10000},
		{name: "read failure", status: 200, readErr: errors.New("read failed")},
		{name: "close failure", status: 200, closeErr: errors.New("close failed")},
		{name: "both failures", status: 200, readErr: errors.New("read failed"), closeErr: errors.New("close failed")},
		{name: "rejected read failure", status: 503, readErr: errors.New(strings.Repeat("read failed", 1000))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var reader io.Reader = strings.NewReader(strings.Repeat("x", tc.size))
			if tc.readErr != nil {
				reader = errorReader{tc.readErr}
			}
			body := &trackedBody{reader: reader, closeErr: tc.closeErr}
			calls := 0
			c := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return response(tc.status, body, "0"), nil
			})}
			got, err := Fetch(context.Background(), c, "http://example.test")
			wantSuccess := tc.status == 200 && tc.size <= 4096 && tc.readErr == nil && tc.closeErr == nil
			if wantSuccess {
				if err != nil || string(got) != strings.Repeat("x", tc.size) {
					t.Errorf("Fetch = %q, %v", got, err)
				}
			} else if got != nil || err == nil || len(err.Error()) > 4096 {
				t.Errorf("Fetch result length = %d, error = %v", len(got), err)
			}
			for _, cause := range []error{tc.readErr, tc.closeErr} {
				if cause != nil && !errors.Is(err, cause) {
					t.Errorf("error does not expose %v", cause)
				}
			}
			if tc.status != 200 && (err == nil || !strings.Contains(err.Error(), fmt.Sprintf("HTTP %d", tc.status))) {
				t.Errorf("missing HTTP rejection: %v", err)
			}
			if body.read > 4097 || body.closes != 1 || calls != 1 {
				t.Errorf("read = %d, closes = %d, attempts = %d", body.read, body.closes, calls)
			}
		})
	}
}

func TestFetchTransportFailureIsTerminalAndBounded(t *testing.T) {
	cause := errors.New(strings.Repeat("transport failed", 1000))
	calls := 0
	c := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return nil, cause
	})}
	_, err := Fetch(context.Background(), c, "http://example.test")
	if !errors.Is(err, cause) || len(err.Error()) > 4096 || calls != 1 {
		t.Errorf("transport error bounded = %v, cause exposed = %v, attempts = %d", len(err.Error()) <= 4096, errors.Is(err, cause), calls)
	}
}

func TestFetchHonorsRedirectPolicyAndBodyOwnership(t *testing.T) {
	policyErr := errors.New("redirect rejected")
	for _, policy := range []error{policyErr, http.ErrUseLastResponse} {
		t.Run(policy.Error(), func(t *testing.T) {
			body := &trackedBody{reader: strings.NewReader("redirect")}
			calls, redirects := 0, 0
			c := &http.Client{
				Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
					calls++
					resp := response(302, body, "")
					resp.Header.Set("Location", "/next")
					return resp, nil
				}),
				CheckRedirect: func(r *http.Request, via []*http.Request) error {
					redirects++
					return policy
				},
			}
			_, err := Fetch(context.Background(), c, "http://example.test")
			if policy == policyErr && !errors.Is(err, policyErr) {
				t.Errorf("lost redirect policy error: %v", err)
			}
			if policy == http.ErrUseLastResponse && (err == nil || !strings.Contains(err.Error(), "HTTP 302")) {
				t.Errorf("last response = %v, want HTTP 302", err)
			}
			if calls != 1 || redirects != 1 || body.closes != 1 {
				t.Errorf("calls = %d, redirects = %d, closes = %d", calls, redirects, body.closes)
			}
		})
	}
}

func TestFetchRealBodyDeadline(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	t.Cleanup(s.Close)
	for _, parentLimit := range []time.Duration{0, 40 * time.Millisecond} {
		t.Run(parentLimit.String(), func(t *testing.T) {
			ctx := context.Background()
			if parentLimit != 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, parentLimit)
				defer cancel()
			}
			start := time.Now()
			_, err := Fetch(ctx, s.Client(), s.URL)
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("body error = %v, want deadline", err)
			}
			limit := 250 * time.Millisecond
			if parentLimit != 0 {
				limit = parentLimit
			}
			if elapsed := time.Since(start); elapsed < limit/2 || elapsed > limit+time.Second {
				t.Errorf("elapsed = %v, want body bounded by %v", elapsed, limit)
			}
		})
	}
}
