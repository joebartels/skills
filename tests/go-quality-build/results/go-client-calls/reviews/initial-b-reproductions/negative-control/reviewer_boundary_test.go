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

// The module's existing test checks Client.Timeout at the transport boundary
// with a context error. This independently checks retention of another cause
// when net/http replaces the returned error with its timeout diagnostic.
func TestReviewerClientTimeoutRetainsIndependentTransportCause(t *testing.T) {
	cause := errors.New("independent transport cause")
	calls := 0
	client := &http.Client{Timeout: 10 * time.Millisecond, Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		defer r.Body.Close()
		calls++
		<-r.Context().Done()
		return nil, errors.Join(cause, r.Context().Err())
	})}
	_, err := Submit(context.Background(), client, "http://example.test", "", nil, false)
	if !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, cause) || calls != 1 {
		t.Fatalf("error=%v; independent cause=%v; calls=%d", err, errors.Is(err, cause), calls)
	}
	t.Logf("calls=%d; deadline=%v; independent cause=%v", calls, errors.Is(err, context.DeadlineExceeded), errors.Is(err, cause))
}

// io.Reader may return data together with an error. Verify that partial data
// cannot turn an incomplete successful response into success, and that a
// terminal status remains terminal even when its diagnostic read fails.
func TestReviewerPartialBodyFailure(t *testing.T) {
	cause := errors.New("partial response failed")
	for _, tc := range []struct {
		name         string
		status       int
		deduplicates bool
		calls        int
	}{
		{"deduplicated success", 200, true, 3},
		{"unprotected success", 200, false, 1},
		{"terminal status", 400, true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var bodies []*observedBody
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				r.Body.Close()
				body := &observedBody{Reader: readerFunc(func(p []byte) (int, error) {
					return copy(p, "partial"), cause
				})}
				bodies = append(bodies, body)
				return response(tc.status, body), nil
			})}
			got, err := Submit(context.Background(), client, "http://example.test", "op", []byte("request"), tc.deduplicates)
			if got != nil || !errors.Is(err, cause) || len(bodies) != tc.calls {
				t.Fatalf("body=%q; error=%v; calls=%d", got, err, len(bodies))
			}
			if tc.status == 400 && !strings.Contains(err.Error(), "HTTP 400: partial") {
				t.Fatalf("terminal diagnostic=%v", err)
			}
			for _, body := range bodies {
				if body.closes != 1 {
					t.Fatalf("closes=%d", body.closes)
				}
			}
			t.Logf("calls=%d; retained cause=%v; all bodies closed", len(bodies), errors.Is(err, cause))
		})
	}
}

func TestReviewerCanceledTerminalDiagnostic(t *testing.T) {
	cause := errors.New("caller canceled terminal diagnostic")
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	calls := 0
	body := &observedBody{Reader: readerFunc(func([]byte) (int, error) {
		cancel(cause)
		return 0, context.Canceled
	})}
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		r.Body.Close()
		calls++
		return response(400, body), nil
	})}
	_, err := Submit(ctx, client, "http://example.test", "op", nil, true)
	if !errors.Is(err, context.Canceled) || !errors.Is(err, cause) || !strings.Contains(err.Error(), "HTTP 400") || calls != 1 || body.closes != 1 {
		t.Fatalf("error=%v; calls=%d; closes=%d", err, calls, body.closes)
	}
	t.Logf("calls=%d; status retained; cancellation and cause retained; closes=%d", calls, body.closes)
}

func TestReviewerEmptyResponse(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		r.Body.Close()
		if r.ContentLength != 0 || r.Header.Get("Idempotency-Key") != "op" {
			t.Errorf("request content length=%d; identity=%q", r.ContentLength, r.Header.Get("Idempotency-Key"))
		}
		return response(200, io.NopCloser(strings.NewReader(""))), nil
	})}
	got, err := Submit(context.Background(), client, "http://example.test", "op", nil, true)
	if err != nil || len(got) != 0 {
		t.Fatalf("body=%q; error=%v", got, err)
	}
}
