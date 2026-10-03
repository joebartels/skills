package stages

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type observedBody struct {
	reader io.Reader
	close  func() error
	closes atomic.Int32
}

func (b *observedBody) Read(p []byte) (int, error) { return b.reader.Read(p) }
func (b *observedBody) Close() error {
	b.closes.Add(1)
	if b.close != nil {
		return b.close()
	}
	return nil
}

func response(r *http.Request, status int, body io.ReadCloser) *http.Response {
	return &http.Response{StatusCode: status, Body: body, Header: make(http.Header), Request: r}
}

func TestSequentialBodies(t *testing.T) {
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		return response(r, http.StatusOK, io.NopCloser(strings.NewReader(r.URL.Path))), nil
	})}
	got, err := FetchAll(context.Background(), client, []string{"http://example.test/first", "http://example.test/second"}, time.Second, time.Second)
	if err != nil || len(got) != 2 || string(got[0]) != "/first" || string(got[1]) != "/second" {
		t.Fatalf("bodies=%q err=%v", got, err)
	}
}

func TestEmptyEndpointsAndAlreadyCanceled(t *testing.T) {
	var calls atomic.Int32
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		return response(r, http.StatusOK, io.NopCloser(strings.NewReader("unexpected"))), nil
	})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := FetchAll(ctx, client, nil, time.Second, time.Second); err != nil || len(got) != 0 {
		t.Fatalf("empty call = %q, %v; want empty success", got, err)
	}
	if _, err := FetchAll(ctx, client, []string{"http://example.test"}, time.Second, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled call error = %v, want context.Canceled", err)
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("transport calls = %d, want 0", got)
	}
}

func TestFailureRetainsCompletedPrefixAndClosesBodies(t *testing.T) {
	readErr, closeErr := errors.New("read failed"), errors.New("close failed")
	var bodies []*observedBody
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/bad" {
			b := &observedBody{reader: errorReader{readErr}, close: func() error { return closeErr }}
			bodies = append(bodies, b)
			return response(r, http.StatusOK, b), nil
		}
		b := &observedBody{reader: strings.NewReader("accepted")}
		bodies = append(bodies, b)
		return response(r, http.StatusOK, b), nil
	})}
	got, err := FetchAll(context.Background(), client, []string{"http://example.test/good", "http://example.test/bad", "http://example.test/later"}, time.Second, time.Second)
	if !errors.Is(err, readErr) || !errors.Is(err, closeErr) {
		t.Fatalf("error = %v, want read and close failures", err)
	}
	if len(got) != 1 || string(got[0]) != "accepted" {
		t.Fatalf("bodies = %q, want completed prefix", got)
	}
	if len(bodies) != 2 || bodies[0].closes.Load() != 1 || bodies[1].closes.Load() != 1 {
		t.Fatalf("returned body close counts = %v, want [1 1]", closeCounts(bodies))
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

func closeCounts(bodies []*observedBody) []int32 {
	counts := make([]int32, len(bodies))
	for i, body := range bodies {
		counts[i] = body.closes.Load()
	}
	return counts
}

func TestNon200ClosesBodyAndRetainsPrefix(t *testing.T) {
	closeErr := errors.New("close failed")
	var badBody *observedBody
	var calls atomic.Int32
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		if r.URL.Path == "/partial" {
			badBody = &observedBody{reader: strings.NewReader("ignored"), close: func() error { return closeErr }}
			return response(r, http.StatusPartialContent, badBody), nil
		}
		return response(r, http.StatusOK, io.NopCloser(strings.NewReader("prefix"))), nil
	})}
	got, err := FetchAll(context.Background(), client, []string{"http://example.test/ok", "http://example.test/partial", "http://example.test/later"}, time.Second, time.Second)
	if !errors.Is(err, closeErr) || !strings.Contains(err.Error(), "status 206") {
		t.Fatalf("error = %v, want status and close errors", err)
	}
	if len(got) != 1 || string(got[0]) != "prefix" || calls.Load() != 2 || badBody.closes.Load() != 1 {
		t.Fatalf("bodies=%q calls=%d bad closes=%d", got, calls.Load(), badBody.closes.Load())
	}
}

func TestStageDeadlineIncludesBodyConsumption(t *testing.T) {
	started := make(chan struct{})
	finished := make(chan struct{})
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		return response(r, http.StatusOK, &contextBody{ctx: r.Context(), started: started, finished: finished}), nil
	})}
	begin := time.Now()
	_, err := FetchAll(context.Background(), client, []string{"http://example.test/slow"}, time.Second, 40*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want stage deadline", err)
	}
	select {
	case <-started:
	default:
		t.Fatal("body was not consumed")
	}
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("body read did not observe stage cancellation")
	}
	if elapsed := time.Since(begin); elapsed > time.Second {
		t.Fatalf("stage did not return within bound: %v", elapsed)
	}
}

func TestCallerCancellationCauseIsRetainedDuringRead(t *testing.T) {
	cause := errors.New("caller stopped operation")
	ctx, cancel := context.WithCancelCause(context.Background())
	started := make(chan struct{})
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		return response(r, http.StatusOK, &contextBody{ctx: r.Context(), started: started, finished: make(chan struct{})}), nil
	})}
	done := make(chan error, 1)
	go func() {
		_, err := FetchAll(ctx, client, []string{"http://example.test/slow"}, time.Second, time.Second)
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request body read did not start")
	}
	cancel(cause)
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
			t.Fatalf("error = %v, want cancellation and caller cause", err)
		}
	case <-time.After(time.Second):
		t.Fatal("FetchAll did not finish after caller cancellation")
	}
}

func TestCancellationAfterCompleteBodyDoesNotRevokeSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		body := &observedBody{reader: strings.NewReader("complete"), close: func() error {
			cancel()
			return nil
		}}
		return response(r, http.StatusOK, body), nil
	})}
	got, err := FetchAll(ctx, client, []string{"http://example.test/complete"}, time.Second, time.Second)
	if err != nil || len(got) != 1 || string(got[0]) != "complete" {
		t.Fatalf("bodies=%q err=%v, want completed success", got, err)
	}
}

type contextBody struct {
	ctx               context.Context
	started, finished chan struct{}
}

func (b *contextBody) Read([]byte) (int, error) {
	close(b.started)
	<-b.ctx.Done()
	close(b.finished)
	return 0, b.ctx.Err()
}
func (*contextBody) Close() error { return nil }

func TestTotalDeadlineIsNotRestartedBetweenStages(t *testing.T) {
	var calls atomic.Int32
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		return response(r, http.StatusOK, &contextBody{ctx: r.Context(), started: make(chan struct{}), finished: make(chan struct{})}), nil
	})}
	_, err := FetchAll(context.Background(), client, []string{"http://example.test/first", "http://example.test/second"}, 50*time.Millisecond, time.Second)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want total deadline", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("transport calls = %d, want only first stage", got)
	}
}
