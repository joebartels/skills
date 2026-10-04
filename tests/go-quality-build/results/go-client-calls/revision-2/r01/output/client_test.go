package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestExistingSuccess(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) }))
	defer s.Close()
	got, err := Submit(context.Background(), s.Client(), s.URL, "op-1", []byte("x"), true)
	if err != nil || string(got) != "ok" {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestSubmitReplay(t *testing.T) {
	payload := []byte{'x', 0, 255, '\n'}
	var mu sync.Mutex
	requests, effects := 0, 0
	var wireErrors []string
	seen := make(map[string][]byte)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		mu.Lock()
		defer mu.Unlock()
		requests++
		key := r.Header.Get("Idempotency-Key")
		if err != nil || r.Method != http.MethodPost || key != "operation-1" || !bytes.Equal(body, payload) {
			wireErrors = append(wireErrors, fmt.Sprintf("method %s, key %q, body %q, error %v", r.Method, key, body, err))
		}
		if previous, ok := seen[key]; ok {
			if !bytes.Equal(previous, body) {
				wireErrors = append(wireErrors, "identity reused with different body")
			}
		} else {
			seen[key] = bytes.Clone(body)
			effects++
		}
		if requests == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("original result"))
	}))
	t.Cleanup(s.Close)
	for i := 0; i < 2; i++ {
		got, err := Submit(context.Background(), s.Client(), s.URL, "operation-1", payload, true)
		if err != nil || string(got) != "original result" {
			t.Fatalf("invocation %d: got %q, %v", i+1, got, err)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if requests != 3 || effects != 1 || len(wireErrors) != 0 {
		t.Fatalf("requests %d, effects %d, wire errors %v", requests, effects, wireErrors)
	}
}

func TestSubmitLostAcknowledgement(t *testing.T) {
	for _, deduplicates := range []bool{false, true} {
		t.Run(fmt.Sprintf("deduplicates=%t", deduplicates), func(t *testing.T) {
			var mu sync.Mutex
			requests, effects := 0, 0
			var wireError string
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				mu.Lock()
				defer mu.Unlock()
				requests++
				key := ""
				if deduplicates {
					key = "committed-operation"
				}
				if err != nil || r.Header.Get("Idempotency-Key") != key || string(body) != "committed payload" {
					wireError = fmt.Sprintf("key %q, body %q, error %v", r.Header.Get("Idempotency-Key"), body, err)
				}
				if requests == 1 {
					effects++
					conn, _, err := w.(http.Hijacker).Hijack()
					if err != nil {
						wireError = err.Error()
						return
					}
					conn.Close() // Commit succeeded, but no acknowledgement reaches the client.
					return
				}
				w.Write([]byte("original result"))
			}))
			t.Cleanup(s.Close)
			got, err := Submit(context.Background(), s.Client(), s.URL, "committed-operation", []byte("committed payload"), deduplicates)
			if deduplicates {
				if err != nil || string(got) != "original result" {
					t.Fatalf("replayed result %q, %v", got, err)
				}
			} else if got != nil || !errors.Is(err, io.EOF) {
				t.Fatalf("ambiguous acknowledgement result %q, %v", got, err)
			}
			mu.Lock()
			defer mu.Unlock()
			wantRequests := 1
			if deduplicates {
				wantRequests = 2
			}
			if requests != wantRequests || effects != 1 || wireError != "" {
				t.Fatalf("requests %d, effects %d, wire error %q", requests, effects, wireError)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackedBody struct {
	reader io.Reader
	read   int
	closes int
}

func (b *trackedBody) Read(p []byte) (int, error) {
	n, err := b.reader.Read(p)
	b.read += n
	return n, err
}

func (b *trackedBody) Close() error { b.closes++; return nil }

func response(r *http.Request, status int, body io.ReadCloser) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: body, Request: r}
}

func TestSubmitAttemptPolicy(t *testing.T) {
	failure := errors.New("lost acknowledgement; effect may have occurred")
	for _, tc := range []struct {
		name         string
		status       int
		deduplicates bool
		wantCalls    int
	}{
		{"503 replay", 503, true, 3},
		{"503 no replay", 503, false, 1},
		{"transport replay", 0, true, 3},
		{"transport no replay", 0, false, 1},
		{"201 terminal", 201, true, 1},
		{"206 terminal", 206, true, 1},
		{"400 terminal", 400, true, 1},
		{"500 terminal", 500, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			var bodies []*trackedBody
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				for _, b := range bodies {
					if b.closes != 1 {
						t.Errorf("previous body closed %d times before retry", b.closes)
					}
				}
				calls++
				wantKey := ""
				if tc.deduplicates {
					wantKey = "stable-identity"
				}
				if got := r.Header.Get("Idempotency-Key"); got != wantKey {
					t.Errorf("idempotency key = %q, want %q", got, wantKey)
				}
				body, err := io.ReadAll(r.Body)
				r.Body.Close()
				if err != nil || string(body) != "unchanged payload" {
					t.Errorf("request body %q, error %v", body, err)
				}
				if tc.status == 0 {
					return nil, failure
				}
				b := &trackedBody{reader: strings.NewReader("diagnostic")}
				bodies = append(bodies, b)
				return response(r, tc.status, b), nil
			})}
			got, err := Submit(context.Background(), client, "http://service.invalid", "stable-identity", []byte("unchanged payload"), tc.deduplicates)
			if got != nil || err == nil || calls != tc.wantCalls {
				t.Fatalf("result %q, error %v, calls %d; want failure and %d calls", got, err, calls, tc.wantCalls)
			}
			if tc.status == 0 && !errors.Is(err, failure) {
				t.Errorf("lost transport cause: %v", err)
			}
			if tc.status != 0 && !strings.Contains(err.Error(), fmt.Sprintf("HTTP %d", tc.status)) {
				t.Errorf("lost status: %v", err)
			}
			for _, b := range bodies {
				if b.closes != 1 {
					t.Errorf("response closed %d times", b.closes)
				}
			}
		})
	}
}

func TestSubmitValidationAndStoppedCaller(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return response(r, 200, io.NopCloser(strings.NewReader("ok"))), nil
	})}
	if _, err := Submit(context.Background(), client, "http://service.invalid", "", nil, true); err == nil {
		t.Fatal("accepted empty operation identity")
	}
	ctx, cancel := context.WithCancelCause(context.Background())
	cause := errors.New("caller stopped")
	cancel(cause)
	if _, err := Submit(ctx, client, "http://service.invalid", "id", nil, true); !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
		t.Fatalf("cancellation causes lost: %v", err)
	}
	if calls != 0 {
		t.Fatalf("validation/stopped caller made %d calls", calls)
	}
	if got, err := Submit(context.Background(), client, "http://service.invalid", "", nil, false); err != nil || string(got) != "ok" {
		t.Fatalf("identity-free single attempt: %q, %v", got, err)
	}
}

func TestSubmitBodyBounds(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		size     int
		wantRead int
		wantOK   bool
	}{
		{"exact success", 200, 4096, 4096, true},
		{"oversized success", 200, 8192, 4097, false},
		{"bounded diagnostic", 400, 8192, 4096, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackedBody{reader: strings.NewReader(strings.Repeat("x", tc.size))}
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return response(r, tc.status, body), nil
			})}
			got, err := Submit(context.Background(), client, "http://service.invalid", "id", nil, true)
			if (err == nil) != tc.wantOK || (tc.wantOK && len(got) != tc.size) || (!tc.wantOK && got != nil) {
				t.Errorf("result length %d, error %v", len(got), err)
			}
			if calls != 1 || body.read != tc.wantRead || body.closes != 1 {
				t.Errorf("calls %d, bytes read %d, closes %d", calls, body.read, body.closes)
			}
		})
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

func TestSubmitReadFailure(t *testing.T) {
	cause := errors.New("response interrupted")
	for _, status := range []int{200, 400} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			body := &trackedBody{reader: errorReader{cause}}
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return response(r, status, body), nil
			})}
			if _, err := Submit(context.Background(), client, "http://service.invalid", "id", nil, false); !errors.Is(err, cause) {
				t.Fatalf("read cause lost: %v", err)
			} else if status != 200 && !strings.Contains(err.Error(), "HTTP 400") {
				t.Fatalf("status lost: %v", err)
			}
			if calls != 1 || body.closes != 1 {
				t.Fatalf("calls %d, closes %d", calls, body.closes)
			}
		})
	}
}

func TestSubmitRedirectIsTerminal(t *testing.T) {
	for _, status := range []int{302, 307} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var targetCalls atomic.Int32
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/target" {
					targetCalls.Add(1)
					w.Write([]byte("ok"))
					return
				}
				w.Header().Set("Location", "/target")
				w.WriteHeader(status)
			}))
			t.Cleanup(s.Close)
			if _, err := Submit(context.Background(), s.Client(), s.URL, "id", []byte("payload"), true); err == nil || !strings.Contains(err.Error(), fmt.Sprintf("HTTP %d", status)) {
				t.Fatalf("redirect result: %v", err)
			}
			if got := targetCalls.Load(); got != 0 {
				t.Fatalf("redirect target received %d calls", got)
			}
		})
	}
}

func TestSubmitPreservesClientPolicy(t *testing.T) {
	policyErr := errors.New("host redirect denied")
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	body := &trackedBody{reader: strings.NewReader("redirect")}
	calls, policyCalls := 0, 0
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		resp := response(r, 307, body)
		resp.Header.Set("Location", "/elsewhere")
		return resp, nil
	})
	client := &http.Client{Transport: transport, Timeout: time.Second, Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error {
		policyCalls++
		return policyErr
	}}
	if _, err := Submit(context.Background(), client, "http://service.invalid", "id", nil, true); !errors.Is(err, policyErr) || !strings.Contains(err.Error(), "HTTP 307") {
		t.Fatalf("policy status/cause lost: %v", err)
	}
	if calls != 1 || policyCalls != 1 || body.closes != 1 || client.Timeout != time.Second || client.Jar != jar {
		t.Fatalf("calls %d, policy calls %d, closes %d, timeout %s, jar preserved %t", calls, policyCalls, body.closes, client.Timeout, client.Jar == jar)
	}
	if client.Transport == nil || client.CheckRedirect == nil {
		t.Fatal("supplied client configuration replaced")
	}
}

func TestSubmitSharesDeadlineAcrossAttempts(t *testing.T) {
	start := time.Now()
	var deadlines []time.Time
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		deadline, ok := r.Context().Deadline()
		if !ok {
			t.Error("request has no deadline")
		}
		deadlines = append(deadlines, deadline)
		return response(r, 503, io.NopCloser(strings.NewReader("busy"))), nil
	})}
	Submit(context.Background(), client, "http://service.invalid", "id", nil, true)
	if len(deadlines) != 3 {
		t.Fatalf("attempts %d, want 3", len(deadlines))
	}
	if delta := deadlines[0].Sub(start); delta < 450*time.Millisecond || delta > 550*time.Millisecond {
		t.Errorf("total deadline allowance %s, want 500ms", delta)
	}
	for _, deadline := range deadlines[1:] {
		if !deadline.Equal(deadlines[0]) {
			t.Errorf("attempt deadline restarted: %s then %s", deadlines[0], deadline)
		}
	}
}

func TestSubmitBudgetIncludesResponseBody(t *testing.T) {
	for _, tc := range []struct {
		name   string
		budget time.Duration
	}{
		{"total budget", 500 * time.Millisecond},
		{"earlier caller deadline", 100 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			started := make(chan struct{}, 1)
			release := make(chan struct{})
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.(http.Flusher).Flush()
				started <- struct{}{}
				select {
				case <-r.Context().Done():
				case <-release:
				}
			}))
			ctx, cancel := context.WithTimeout(context.Background(), tc.budget)
			if tc.budget == 500*time.Millisecond {
				cancel()
				ctx, cancel = context.WithCancel(context.Background())
			}
			result := make(chan error, 1)
			finished := make(chan struct{})
			t.Cleanup(func() {
				cancel()
				close(release)
				select {
				case <-finished:
				case <-time.After(2 * time.Second):
					t.Error("Submit did not finish during cleanup")
				}
				s.Close()
			})
			start := time.Now()
			go func() {
				defer close(finished)
				_, err := Submit(ctx, s.Client(), s.URL, "id", nil, true)
				result <- err
			}()
			select {
			case <-started:
			case err := <-result:
				t.Fatalf("returned before response started: %v", err)
			case <-time.After(2 * time.Second):
				t.Fatal("request did not start")
			}
			select {
			case err := <-result:
				elapsed := time.Since(start)
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("body deadline cause lost: %v", err)
				}
				if elapsed < tc.budget-40*time.Millisecond || elapsed > tc.budget+500*time.Millisecond {
					t.Errorf("returned after %s, budget %s", elapsed, tc.budget)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("response body exceeded its budget")
			}
		})
	}
}

func TestSubmitCancellationPreventsRetryAndPreservesFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	cause := errors.New("independent transport failure")
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		cancel()
		return nil, cause
	})}
	_, err := Submit(ctx, client, "http://service.invalid", "id", nil, true)
	if !errors.Is(err, cause) || !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("error %v, calls %d", err, calls)
	}
}

type legacyTransport struct {
	started chan struct{}
	stopped chan struct{}
	cause   error
	body    bool
	closes  atomic.Int32
	stop    sync.Once
}

func (t *legacyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.body {
		return response(r, 200, &legacyBody{transport: t}), nil
	}
	close(t.started)
	<-t.stopped
	return nil, t.cause
}

func (t *legacyTransport) CancelRequest(*http.Request) {
	t.stop.Do(func() { close(t.stopped) })
}

type legacyBody struct{ transport *legacyTransport }

func (b *legacyBody) Read([]byte) (int, error) {
	close(b.transport.started)
	<-b.transport.stopped
	return 0, b.transport.cause
}

func (b *legacyBody) Close() error { b.transport.closes.Add(1); return nil }

func submitWithLegacyTransport(t *testing.T, ctx context.Context, client *http.Client, transport *legacyTransport) error {
	t.Helper()
	finished := make(chan struct{})
	result := make(chan error, 1)
	t.Cleanup(func() {
		transport.CancelRequest(nil)
		select {
		case <-finished:
		case <-time.After(2 * time.Second):
			t.Error("legacy transport call did not finish during cleanup")
		}
	})
	go func() {
		defer close(finished)
		_, err := Submit(ctx, client, "http://service.invalid", "id", nil, false)
		result <- err
	}()
	select {
	case err := <-result:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("legacy transport call did not stop")
		return nil
	}
}

func TestSubmitPreservesClientTimeoutAndOriginalCause(t *testing.T) {
	for _, body := range []bool{false, true} {
		t.Run(fmt.Sprintf("body=%t", body), func(t *testing.T) {
			cause := errors.New("independent failure during client timeout")
			transport := &legacyTransport{started: make(chan struct{}), stopped: make(chan struct{}), cause: cause, body: body}
			client := &http.Client{Transport: transport, Timeout: 20 * time.Millisecond}
			err := submitWithLegacyTransport(t, context.Background(), client, transport)
			if !errors.Is(err, cause) || !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("timeout/independent cause lost: %v", err)
			}
			select {
			case <-transport.started:
			default:
				t.Fatal("transport/body was never started")
			}
			select {
			case <-transport.stopped:
			default:
				t.Fatal("legacy client timeout cancellation was lost")
			}
			wantCloses := int32(0)
			if body {
				wantCloses = 1
			}
			if client.Timeout != 20*time.Millisecond || client.Transport != transport || transport.closes.Load() != wantCloses {
				t.Fatalf("timeout %s, transport preserved %t, closes %d", client.Timeout, client.Transport == transport, transport.closes.Load())
			}
		})
	}
}

func TestSubmitOperationBudgetWithLegacyTransport(t *testing.T) {
	for _, timeout := range []time.Duration{0, time.Second} {
		t.Run(timeout.String(), func(t *testing.T) {
			cause := errors.New("transport stopped")
			transport := &legacyTransport{started: make(chan struct{}), stopped: make(chan struct{}), cause: cause}
			client := &http.Client{Transport: transport, Timeout: timeout}
			start := time.Now()
			err := submitWithLegacyTransport(t, context.Background(), client, transport)
			if elapsed := time.Since(start); elapsed < 450*time.Millisecond || elapsed > time.Second {
				t.Errorf("operation budget elapsed %s, want 500ms", elapsed)
			}
			if !errors.Is(err, cause) || !errors.Is(err, context.DeadlineExceeded) || client.Timeout != timeout {
				t.Fatalf("error %v, supplied timeout %s", err, client.Timeout)
			}
		})
	}
}

func TestSubmitCallerCancellationWithLegacyTransport(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	cause := errors.New("transport stopped")
	transport := &legacyTransport{started: make(chan struct{}), stopped: make(chan struct{}), cause: cause}
	client := &http.Client{Transport: transport, Timeout: time.Second}
	canceled := make(chan struct{})
	go func() {
		defer close(canceled)
		select {
		case <-transport.started:
			cancel()
		case <-ctx.Done():
		}
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-canceled:
		case <-time.After(2 * time.Second):
			t.Error("cancellation observer did not finish")
		}
	})
	start := time.Now()
	err := submitWithLegacyTransport(t, ctx, client, transport)
	if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
		t.Errorf("caller cancellation took %s", elapsed)
	}
	if !errors.Is(err, cause) || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation causes lost: %v", err)
	}
}
