package client

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var _ func(context.Context, *http.Client, string) ([]byte, error) = Fetch

func TestExistingSuccess(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) }))
	defer s.Close()
	got, err := Fetch(context.Background(), s.Client(), s.URL)
	if err != nil || string(got) != "ok" {
		t.Fatalf("got %q, %v", got, err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackedBody struct {
	reader   io.Reader
	read     int
	closed   int
	closeErr error
}

func (b *trackedBody) Read(p []byte) (int, error) {
	n, err := b.reader.Read(p)
	b.read += n
	return n, err
}

func (b *trackedBody) Close() error { b.closed++; return b.closeErr }

func TestStatusesAndAttemptLimit(t *testing.T) {
	for _, tc := range []struct {
		name     string
		statuses []int
		calls    int
		wantErr  string
	}{
		{"retryable then success", []int{429, 503, 200}, 3, ""},
		{"attempt limit", []int{503, 429, 503, 200}, 3, "HTTP 503"},
		{"partial content is terminal", []int{206, 200}, 1, "HTTP 206"},
		{"other server error is terminal", []int{500, 200}, 1, "HTTP 500"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			var bodies []*trackedBody
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if calls > 0 && bodies[calls-1].closed != 1 {
					return nil, errors.New("previous body remains open")
				}
				status := tc.statuses[calls]
				calls++
				body := &trackedBody{reader: strings.NewReader("ok")}
				bodies = append(bodies, body)
				return &http.Response{StatusCode: status, Header: http.Header{"Retry-After": {"0"}}, Body: body}, nil
			})}
			got, err := Fetch(context.Background(), client, "http://example.test")
			if tc.wantErr == "" {
				if err != nil || string(got) != "ok" {
					t.Fatalf("Fetch = %q, %v", got, err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantErr) || got != nil {
				t.Fatalf("Fetch = %q, %v; want %s", got, err, tc.wantErr)
			}
			if calls != tc.calls {
				t.Errorf("calls = %d, want %d", calls, tc.calls)
			}
			for i, body := range bodies {
				if body.closed != 1 {
					t.Errorf("body %d closed %d times", i, body.closed)
				}
			}
		})
	}
}

func TestBodyBounds(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		size   int
	}{
		{"exact success limit", 200, 4096},
		{"oversized success", 200, 8000},
		{"bounded rejection", 400, 8000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackedBody{reader: strings.NewReader(strings.Repeat("x", tc.size))}
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: body, Header: make(http.Header)}, nil
			})}
			got, err := Fetch(context.Background(), client, "http://example.test")
			if tc.status == 200 && tc.size == 4096 {
				if err != nil || len(got) != 4096 {
					t.Fatalf("Fetch returned %d bytes, %v", len(got), err)
				}
			} else {
				if err == nil || got != nil {
					t.Fatalf("Fetch returned %d bytes, %v; want error", len(got), err)
				}
				if len(err.Error()) > 4096 {
					t.Errorf("diagnostic size = %d", len(err.Error()))
				}
				if tc.status == 400 && !strings.Contains(err.Error(), "HTTP 400") {
					t.Errorf("missing rejection: %v", err)
				}
			}
			wantRead := tc.size
			if wantRead > 4097 {
				wantRead = 4097
			}
			if body.read != wantRead || body.closed != 1 {
				t.Errorf("body read=%d closed=%d, want read=%d closed=1", body.read, body.closed, wantRead)
			}
		})
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

func TestFailuresRemainDiscoverable(t *testing.T) {
	for _, status := range []int{200, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			cause := errors.New(strings.Repeat("read failure ", 1000))
			body := &trackedBody{reader: errorReader{errors.Join(context.Canceled, cause)}}
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: status, Body: body, Header: make(http.Header)}, nil
			})}
			_, err := Fetch(context.Background(), client, "http://example.test")
			if !errors.Is(err, cause) || !errors.Is(err, context.Canceled) {
				t.Fatalf("lost read causes: %v", err)
			}
			if len(err.Error()) > 4096 || calls != 1 || body.closed != 1 {
				t.Errorf("diagnostic size=%d calls=%d closes=%d", len(err.Error()), calls, body.closed)
			}
			if status == 503 && !strings.Contains(err.Error(), "HTTP 503") {
				t.Errorf("lost HTTP rejection: %v", err)
			}
		})
	}
	transportErr := errors.New(strings.Repeat("transport failure ", 1000))
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, transportErr
	})}
	_, err := Fetch(context.Background(), client, "http://example.test")
	if !errors.Is(err, transportErr) || calls != 1 || len(err.Error()) > 4096 {
		t.Fatalf("transport error=%v calls=%d", err, calls)
	}
}

func TestBodyCloseFailure(t *testing.T) {
	for _, status := range []int{200, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			cause := errors.New("close failed")
			body := &trackedBody{reader: strings.NewReader("ok"), closeErr: cause}
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: status, Body: body, Header: make(http.Header)}, nil
			})}
			got, err := Fetch(context.Background(), client, "http://example.test")
			if got != nil || !errors.Is(err, cause) || calls != 1 || body.closed != 1 {
				t.Fatalf("Fetch=%q, %v calls=%d closes=%d", got, err, calls, body.closed)
			}
			if status == 503 && !strings.Contains(err.Error(), "HTTP 503") {
				t.Errorf("lost rejection: %v", err)
			}
		})
	}
}

func TestRetryAfterUnfit(t *testing.T) {
	for _, hint := range []string{"1", "9223372036854775807", strings.Repeat("9", 100), time.Now().Add(time.Hour).UTC().Format(http.TimeFormat)} {
		t.Run(hint[:min(len(hint), 30)], func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": {hint}}, Body: io.NopCloser(strings.NewReader("busy"))}, nil
			})}
			start := time.Now()
			_, err := Fetch(context.Background(), client, "http://example.test")
			if err == nil || !strings.Contains(err.Error(), "HTTP 429") || calls != 1 {
				t.Fatalf("error=%v calls=%d", err, calls)
			}
			if time.Since(start) > 150*time.Millisecond {
				t.Errorf("unfit hint waited %v", time.Since(start))
			}
		})
	}
}

func TestRetryAfterFallbackAndPastDate(t *testing.T) {
	for _, tc := range []struct {
		name string
		hint string
		wait bool
	}{
		{"malformed", "nonsense", true},
		{"negative", "-1", true},
		{"missing", "", true},
		{"past date", "Mon, 02 Jan 2006 15:04:05 GMT", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			var first time.Time
			var gap time.Duration
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				status := 200
				if calls == 1 {
					first = time.Now()
					status = 503
				} else {
					gap = time.Since(first)
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Retry-After": {tc.hint}}, Body: io.NopCloser(strings.NewReader("ok"))}, nil
			})}
			got, err := Fetch(context.Background(), client, "http://example.test")
			if err != nil || string(got) != "ok" || calls != 2 {
				t.Fatalf("Fetch=%q, %v calls=%d", got, err, calls)
			}
			if tc.wait && gap < 10*time.Millisecond {
				t.Errorf("fallback retried after %v", gap)
			}
		})
	}
}

func TestCancellationRetainsLastRejection(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	cause := errors.New("caller stopped")
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		cancel(cause)
		return &http.Response{StatusCode: 503, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("busy"))}, nil
	})}
	_, err := Fetch(ctx, client, "http://example.test")
	if err == nil || !strings.Contains(err.Error(), "HTTP 503") || !errors.Is(err, context.Canceled) || !errors.Is(err, cause) || calls != 1 {
		t.Fatalf("error=%v calls=%d", err, calls)
	}
}

func TestFutureRetryDate(t *testing.T) {
	for _, cancelDuringWait := range []bool{false, true} {
		name := "wait until date"
		if cancelDuringWait {
			name = "cancel during wait"
		}
		t.Run(name, func(t *testing.T) {
			// HTTP dates have whole-second resolution. Start 150ms before one
			// so a future hint fits the operation budget.
			now := time.Now()
			date := now.Truncate(time.Second).Add(time.Second)
			startAt := date.Add(-150 * time.Millisecond)
			if startAt.Before(now) {
				date = date.Add(time.Second)
				startAt = startAt.Add(time.Second)
			}
			timer := time.NewTimer(time.Until(startAt))
			<-timer.C
			if time.Until(date) < 75*time.Millisecond {
				t.Skip("scheduler missed the fitting HTTP-date window")
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var cancelTimer *time.Timer
			t.Cleanup(func() {
				if cancelTimer != nil {
					cancelTimer.Stop()
				}
			})
			calls := 0
			var second time.Time
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				status := 503
				if calls == 1 && cancelDuringWait {
					cancelTimer = time.AfterFunc(20*time.Millisecond, cancel)
				}
				if calls == 2 {
					second = time.Now()
					status = 200
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Retry-After": {date.UTC().Format(http.TimeFormat)}}, Body: io.NopCloser(strings.NewReader("ok"))}, nil
			})}
			got, err := Fetch(ctx, client, "http://example.test")
			if cancelDuringWait {
				if !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), "HTTP 503") || calls != 1 {
					t.Fatalf("error=%v calls=%d", err, calls)
				}
				return
			}
			if err != nil || string(got) != "ok" || calls != 2 || second.Before(date) {
				t.Fatalf("Fetch=%q, %v calls=%d second=%v retry date=%v", got, err, calls, second, date)
			}
		})
	}
}

func TestCanceledAtEntry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("unexpected request")
	})}
	_, err := Fetch(ctx, client, "http://example.test")
	if !errors.Is(err, context.Canceled) || calls != 0 {
		t.Fatalf("error=%v calls=%d", err, calls)
	}
}

func TestSharedBudgetIncludesBody(t *testing.T) {
	for _, parentLimit := range []time.Duration{0, 80 * time.Millisecond} {
		t.Run(parentLimit.String(), func(t *testing.T) {
			ctx := context.Background()
			if parentLimit > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, parentLimit)
				defer cancel()
			}
			calls := 0
			var deadlines []time.Time
			var bodies []*trackedBody
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				deadline, ok := r.Context().Deadline()
				if !ok {
					return nil, errors.New("missing request deadline")
				}
				deadlines = append(deadlines, deadline)
				body := &trackedBody{reader: strings.NewReader("busy")}
				if calls == 2 {
					body.reader = contextReader{r.Context()}
				}
				bodies = append(bodies, body)
				return &http.Response{StatusCode: 503, Header: make(http.Header), Body: body}, nil
			})}
			start := time.Now()
			_, err := Fetch(ctx, client, "http://example.test")
			if err == nil || !strings.Contains(err.Error(), "HTTP 503") || !errors.Is(err, context.DeadlineExceeded) || calls != 2 {
				t.Fatalf("error=%v calls=%d", err, calls)
			}
			limit := 250 * time.Millisecond
			if parentLimit > 0 {
				limit = parentLimit
			}
			if elapsed := time.Since(start); elapsed < limit-5*time.Millisecond || elapsed > limit+time.Second {
				t.Errorf("elapsed=%v, expected budget %v", elapsed, limit)
			}
			if !deadlines[0].Equal(deadlines[1]) || deadlines[0].After(start.Add(limit+5*time.Millisecond)) {
				t.Errorf("attempt deadlines = %v", deadlines)
			}
			for i, body := range bodies {
				if body.closed != 1 {
					t.Errorf("body %d closed %d times", i, body.closed)
				}
			}
		})
	}
}

type contextReader struct{ ctx context.Context }

func (r contextReader) Read([]byte) (int, error) {
	<-r.ctx.Done()
	return 0, r.ctx.Err()
}

func TestSuppliedRedirectPolicy(t *testing.T) {
	redirectErr := errors.New("redirect blocked")
	body := &trackedBody{reader: strings.NewReader("moved")}
	policyCalls := 0
	client := &http.Client{
		Timeout:       time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { policyCalls++; return redirectErr },
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 302, Header: http.Header{"Location": {"http://example.test/other"}}, Body: body}, nil
		}),
	}
	_, err := Fetch(context.Background(), client, "http://example.test")
	if !errors.Is(err, redirectErr) || body.closed != 1 || policyCalls != 1 || client.Timeout != time.Second {
		t.Fatalf("error=%v closed=%d policy calls=%d timeout=%v", err, body.closed, policyCalls, client.Timeout)
	}
}

func TestNetworkBodyCancellation(t *testing.T) {
	for _, tc := range []struct {
		name          string
		parentTimeout time.Duration
		clientTimeout time.Duration
		wantDuration  time.Duration
	}{
		{"parent deadline", 50 * time.Millisecond, 0, 50 * time.Millisecond},
		{"client timeout", 0, 30 * time.Millisecond, 30 * time.Millisecond},
		{"total budget", 0, 0, 250 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			finished := make(chan struct{})
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer close(finished)
				w.WriteHeader(200)
				w.(http.Flusher).Flush()
				<-r.Context().Done()
			}))
			defer s.Close()
			ctx := context.Background()
			if tc.parentTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tc.parentTimeout)
				defer cancel()
			}
			client := s.Client()
			client.Timeout = tc.clientTimeout
			start := time.Now()
			_, err := Fetch(ctx, client, s.URL)
			if tc.clientTimeout > 0 {
				var timeoutErr net.Error
				if !errors.As(err, &timeoutErr) || !timeoutErr.Timeout() {
					t.Fatalf("error=%v, want client timeout", err)
				}
			} else if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("error=%v, want context deadline", err)
			}
			if elapsed := time.Since(start); elapsed > tc.wantDuration+time.Second {
				t.Errorf("body cancellation took %v", elapsed)
			}
			if client.Timeout != tc.clientTimeout {
				t.Errorf("client timeout changed to %v", client.Timeout)
			}
			select {
			case <-finished:
			case <-time.After(time.Second):
				t.Fatal("handler did not observe request cancellation")
			}
		})
	}
}
