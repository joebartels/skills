package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sony/gobreaker/v2"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func response(status int) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader("ok"))}
}

func fetchStatus(t *testing.T, c *Client, endpoint string, status int) {
	t.Helper()
	got, err := c.Fetch(context.Background(), endpoint)
	if status == http.StatusOK {
		if err != nil || string(got) != "ok" {
			t.Fatalf("Fetch(%q) = %q, %v; want ok", endpoint, got, err)
		}
	} else if err == nil || err.Error() != fmt.Sprintf("HTTP %d", status) {
		t.Fatalf("Fetch(%q) = %q, %v; want HTTP %d", endpoint, got, err, status)
	}
}

type trackedBody struct {
	io.Reader
	closed  int
	onClose func()
}

func (b *trackedBody) Close() error {
	b.closed++
	if b.onClose != nil {
		b.onClose()
	}
	return nil
}

type readerFunc func([]byte) (int, error)

func (f readerFunc) Read(p []byte) (int, error) { return f(p) }

func TestFetchReadFailureOwnsBodyAndContext(t *testing.T) {
	readErr := errors.New("body read failed")
	type contextKey struct{}
	parent := context.WithValue(context.Background(), contextKey{}, "caller value")
	start := time.Now()
	var requestCtx context.Context
	body := &trackedBody{Reader: readerFunc(func(p []byte) (int, error) {
		p[0] = 'x'
		return 1, readErr
	})}
	body.onClose = func() {
		if requestCtx.Err() != nil {
			t.Error("request scope ended before owned body closure")
		}
	}
	c := New(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requestCtx = r.Context()
		deadline, ok := requestCtx.Deadline()
		if !ok || deadline.Before(start.Add(200*time.Millisecond)) || deadline.After(start.Add(300*time.Millisecond)) {
			t.Errorf("request deadline = %v, %v; want 250ms total budget", deadline, ok)
		}
		if requestCtx.Value(contextKey{}) != "caller value" {
			t.Error("caller context value was lost")
		}
		return &http.Response{StatusCode: 200, Body: body}, nil
	})}, true)
	got, err := c.Fetch(parent, "http://dependency.test/read")
	if string(got) != "x" || !errors.Is(err, readErr) || body.closed != 1 {
		t.Fatalf("Fetch = %q, %v, closes %d; want x, read error, 1", got, err, body.closed)
	}
	if requestCtx.Err() != context.Canceled {
		t.Fatalf("request scope after Fetch = %v; want canceled", requestCtx.Err())
	}
}

func TestFetchStatusAndBodyLimit(t *testing.T) {
	for _, status := range []int{200, 206, 302, 400, 503} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			body := &trackedBody{Reader: strings.NewReader(strings.Repeat("x", 5000))}
			calls := 0
			c := New(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: status, Body: body}, nil
			})}, false)
			got, err := c.Fetch(context.Background(), "http://dependency.test/data")
			if status == 200 {
				if err != nil || string(got) != strings.Repeat("x", 4096) {
					t.Fatalf("Fetch = %d bytes, %v; want 4096 x bytes", len(got), err)
				}
			} else if err == nil || err.Error() != fmt.Sprintf("HTTP %d", status) || len(got) != 0 {
				t.Fatalf("Fetch = %q, %v; want HTTP %d", got, err, status)
			}
			if calls != 1 || body.closed != 1 {
				t.Fatalf("requests = %d, body closes = %d; want 1 each", calls, body.closed)
			}
		})
	}
}

func TestFetchDoesNotFollowRedirectOrMutateClient(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			var requests atomic.Int32
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.URL.Path == "/redirect" {
					http.Redirect(w, r, "/ok", http.StatusFound)
					return
				}
				io.WriteString(w, "ok")
			}))
			t.Cleanup(s.Close)
			borrowed := s.Client()
			borrowed.Timeout = time.Second
			redirects := 0
			borrowed.CheckRedirect = func(*http.Request, []*http.Request) error { redirects++; return nil }
			transport := borrowed.Transport
			c := New(borrowed, enabled)
			fetchStatus(t, c, s.URL+"/redirect", 302)
			if requests.Load() != 1 || redirects != 0 {
				t.Fatalf("requests = %d, redirect callbacks = %d; want 1, 0", requests.Load(), redirects)
			}
			if c.http != borrowed || borrowed.Transport != transport || borrowed.Timeout != time.Second {
				t.Fatal("supplied client or its configuration changed")
			}
			resp, err := borrowed.Get(s.URL + "/redirect")
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()
			if resp.StatusCode != 200 || redirects != 1 || requests.Load() != 3 {
				t.Fatalf("borrowed client lost redirect policy: status %d, callbacks %d, requests %d", resp.StatusCode, redirects, requests.Load())
			}
		})
	}
}

func TestFetchTotalBudget(t *testing.T) {
	for _, tc := range []struct {
		name   string
		body   bool
		parent time.Duration
	}{
		{name: "headers"},
		{name: "body", body: true},
		{name: "earlier_parent", body: true, parent: 40 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			release := make(chan struct{})
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.body {
					w.WriteHeader(http.StatusOK)
					w.(http.Flusher).Flush()
				}
				select {
				case <-r.Context().Done():
				case <-release:
				}
			}))
			t.Cleanup(func() { close(release); s.Close() })
			ctx := context.Background()
			if tc.parent != 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tc.parent)
				defer cancel()
			}
			borrowed := s.Client()
			// Bound the old implementation too, without replacing the Fetch budget.
			borrowed.Timeout = time.Second
			start := time.Now()
			_, err := New(borrowed, true).Fetch(ctx, s.URL)
			elapsed := time.Since(start)
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("Fetch = %v; want deadline exceeded", err)
			}
			minimum, maximum := 200*time.Millisecond, 750*time.Millisecond
			if tc.parent != 0 {
				minimum, maximum = 20*time.Millisecond, 500*time.Millisecond
			}
			if elapsed < minimum || elapsed > maximum {
				t.Fatalf("Fetch duration = %v; want %v..%v", elapsed, minimum, maximum)
			}
		})
	}
}

func TestBreakerDependencyScope(t *testing.T) {
	calls := 0
	borrowed := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path == "/ok" {
			return response(200), nil
		}
		return response(503), nil
	})}
	c := New(borrowed, true)
	fetchStatus(t, c, "http://a.test/one", 503)
	fetchStatus(t, c, "http://b.test/one", 503)
	fetchStatus(t, c, "http://a.test/two?query=yes", 503)
	fetchStatus(t, c, "http://b.test/ok", 200)
	fetchStatus(t, c, "http://a.test/three", 503)
	_, err := c.Fetch(context.Background(), "http://a.test/four")
	if !errors.Is(err, gobreaker.ErrOpenState) || calls != 5 {
		t.Fatalf("same-authority rejection = %v, requests %d; want open, 5", err, calls)
	}
	fetchStatus(t, c, "http://b.test/two", 503)
	fetchStatus(t, c, "https://a.test/ok", 200)
	fetchStatus(t, c, "http://a.test:8080/ok", 200)
	fetchStatus(t, New(borrowed, true), "http://a.test/one", 503)
	if calls != 9 {
		t.Fatalf("requests = %d; want 9 for independent dependencies and clients", calls)
	}
}

func TestBreakerExcludedOutcomesPreserveFailureHistory(t *testing.T) {
	for _, excluded := range []string{"400", "500", "206", "transport", "canceled"} {
		t.Run(excluded, func(t *testing.T) {
			calls := 0
			c := New(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				switch r.URL.Path {
				case "/400":
					return response(400), nil
				case "/500":
					return response(500), nil
				case "/206":
					return response(206), nil
				case "/transport":
					return nil, errors.New("transport failed")
				}
				return response(503), nil
			})}, true)
			for i := 0; i < 2; i++ {
				fetchStatus(t, c, "http://dependency.test/failure", 503)
				ctx := context.Background()
				if excluded == "canceled" {
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(ctx)
					cancel()
				}
				if _, err := c.Fetch(ctx, "http://dependency.test/"+excluded); err == nil {
					t.Fatal("excluded request unexpectedly succeeded")
				}
			}
			fetchStatus(t, c, "http://dependency.test/failure", 503)
			before := calls
			if _, err := c.Fetch(context.Background(), "http://dependency.test/failure"); !errors.Is(err, gobreaker.ErrOpenState) || calls != before {
				t.Fatalf("after three eligible failures: error %v, requests %d -> %d", err, before, calls)
			}
		})
	}
}

func TestBreakerSuccessResetsFailuresAndDisabledNeverTrips(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint(enabled), func(t *testing.T) {
			c := New(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/ok" {
					return response(200), nil
				}
				return response(503), nil
			})}, enabled)
			for _, status := range []int{503, 503, 200, 503, 503, 503} {
				path := "/failure"
				if status == 200 {
					path = "/ok"
				}
				fetchStatus(t, c, "http://dependency.test"+path, status)
			}
			if enabled {
				if _, err := c.Fetch(context.Background(), "http://dependency.test/ok"); !errors.Is(err, gobreaker.ErrOpenState) {
					t.Fatalf("after three new failures: %v; want open", err)
				}
			} else {
				fetchStatus(t, c, "http://dependency.test/failure", 503)
			}
		})
	}
}

func TestBreakerCanceled503DoesNotCountOrReset(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	c := New(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path == "/canceled" {
			cancel()
		}
		return response(503), nil
	})}, true)
	for i := 0; i < 2; i++ {
		fetchStatus(t, c, "http://dependency.test/failure", 503)
	}
	if _, err := c.Fetch(ctx, "http://dependency.test/canceled"); err == nil || err.Error() != "HTTP 503" {
		t.Fatalf("independent 503 during cancellation = %v; want HTTP 503", err)
	}
	fetchStatus(t, c, "http://dependency.test/failure", 503)
	if _, err := c.Fetch(context.Background(), "http://dependency.test/failure"); !errors.Is(err, gobreaker.ErrOpenState) || calls != 4 {
		t.Fatalf("canceled 503 changed history: error %v, requests %d; want open, 4", err, calls)
	}
}

func waitForRecovery() { time.Sleep(110 * time.Millisecond) }

func waitError(t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("Fetch did not finish")
		return nil
	}
}

func TestBreakerRecoveryAdmissionAndExcludedCompletion(t *testing.T) {
	for _, excluded := range []string{"400", "500", "transport", "canceled"} {
		t.Run(excluded, func(t *testing.T) {
			started := make(chan struct{})
			release := make(chan struct{})
			var calls atomic.Int32
			c := New(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls.Add(1)
				if r.URL.Path == "/probe" {
					close(started)
					select {
					case <-release:
					case <-r.Context().Done():
						return nil, r.Context().Err()
					}
					if excluded == "transport" {
						return nil, errors.New("transport failed")
					}
					if excluded == "500" {
						return response(500), nil
					}
					return response(400), nil
				}
				if r.URL.Path == "/ok" {
					return response(200), nil
				}
				return response(503), nil
			})}, true)
			for i := 0; i < 3; i++ {
				fetchStatus(t, c, "http://dependency.test/failure", 503)
			}
			waitForRecovery()
			ctx, cancel := context.WithCancel(context.Background())
			result := make(chan error, 1)
			finished := make(chan struct{})
			t.Cleanup(func() {
				cancel()
				select {
				case <-finished:
				case <-time.After(2 * time.Second):
					t.Error("probe did not finish during cleanup")
				}
			})
			go func() {
				defer close(finished)
				_, err := c.Fetch(ctx, "http://dependency.test/probe")
				result <- err
			}()
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("recovery probe not admitted")
			}
			if _, err := c.Fetch(context.Background(), "http://dependency.test/ok"); !errors.Is(err, gobreaker.ErrTooManyRequests) || calls.Load() != 4 {
				t.Fatalf("concurrent recovery = %v, requests %d; want rejected, 4", err, calls.Load())
			}
			if excluded == "canceled" {
				cancel()
			} else {
				close(release)
			}
			if err := waitError(t, result); err == nil || (excluded == "canceled" && !errors.Is(err, context.Canceled)) {
				t.Fatalf("excluded probe = %v", err)
			}
			fetchStatus(t, c, "http://dependency.test/failure", 503)
			if _, err := c.Fetch(context.Background(), "http://dependency.test/ok"); !errors.Is(err, gobreaker.ErrOpenState) || calls.Load() != 5 {
				t.Fatalf("failed replacement probe = %v, requests %d; want reopened, 5", err, calls.Load())
			}
			waitForRecovery()
			fetchStatus(t, c, "http://dependency.test/ok", 200)
			for i := 0; i < 3; i++ {
				fetchStatus(t, c, "http://dependency.test/failure", 503)
			}
			if _, err := c.Fetch(context.Background(), "http://dependency.test/ok"); !errors.Is(err, gobreaker.ErrOpenState) {
				t.Fatalf("recovered breaker did not close/reset correctly: %v", err)
			}
		})
	}
}

func TestBreakerIgnoresLateGenerationCompletions(t *testing.T) {
	for _, lateStatus := range []int{200, 503} {
		t.Run(fmt.Sprint(lateStatus), func(t *testing.T) {
			started := make(chan struct{})
			release := make(chan struct{})
			c := New(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/late" {
					close(started)
					select {
					case <-release:
						return response(lateStatus), nil
					case <-r.Context().Done():
						return nil, r.Context().Err()
					}
				}
				if r.URL.Path == "/ok" {
					return response(200), nil
				}
				return response(503), nil
			})}, true)
			ctx, cancel := context.WithCancel(context.Background())
			result := make(chan error, 1)
			finished := make(chan struct{})
			t.Cleanup(func() {
				cancel()
				select {
				case <-finished:
				case <-time.After(2 * time.Second):
					t.Error("late request did not finish during cleanup")
				}
			})
			go func() {
				defer close(finished)
				_, err := c.Fetch(ctx, "http://dependency.test/late")
				result <- err
			}()
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("late request not started")
			}
			for i := 0; i < 3; i++ {
				fetchStatus(t, c, "http://dependency.test/failure", 503)
			}
			waitForRecovery()
			fetchStatus(t, c, "http://dependency.test/ok", 200)
			for i := 0; i < 2; i++ {
				fetchStatus(t, c, "http://dependency.test/failure", 503)
			}
			close(release)
			err := waitError(t, result)
			if lateStatus == 200 && err != nil || lateStatus == 503 && (err == nil || err.Error() != "HTTP 503") {
				t.Fatalf("late response = %v; want status %d", err, lateStatus)
			}
			fetchStatus(t, c, "http://dependency.test/failure", 503)
			if _, err := c.Fetch(context.Background(), "http://dependency.test/ok"); !errors.Is(err, gobreaker.ErrOpenState) {
				t.Fatalf("late completion changed current failure history: %v", err)
			}
		})
	}
}
