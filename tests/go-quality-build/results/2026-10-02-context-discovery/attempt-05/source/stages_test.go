package stages

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type trackingBody struct {
	reader   io.Reader
	readErr  error
	closeErr error
	closed   bool
}

func (b *trackingBody) Read(p []byte) (int, error) {
	if b.readErr != nil {
		return 0, b.readErr
	}
	return b.reader.Read(p)
}

func (b *trackingBody) Close() error {
	b.closed = true
	return b.closeErr
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

func TestAlreadyCanceledStartsNoRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("unexpected request")
	})}
	got, err := FetchAll(ctx, client, []string{"http://example.test/"}, time.Second, time.Second)
	if err == nil || calls != 0 || len(got) != 0 {
		t.Fatalf("bodies=%q err=%v calls=%d", got, err, calls)
	}
}

func TestParentDeadlineBoundsEachStageAndPreservesPrefix(t *testing.T) {
	parent, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	calls := 0
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return response(r, http.StatusOK, io.NopCloser(strings.NewReader("first"))), nil
		}
		deadline, ok := r.Context().Deadline()
		if !ok || time.Until(deadline) > 120*time.Millisecond {
			t.Errorf("second stage deadline=%v, want bounded by configured stage", deadline)
		}
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	got, err := FetchAll(parent, client, []string{"http://example.test/first", "http://example.test/second"}, time.Second, 80*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) || len(got) != 1 || string(got[0]) != "first" || calls != 2 {
		t.Fatalf("bodies=%q err=%v calls=%d", got, err, calls)
	}
}

func TestTotalBudgetIsNotRestartedAndParentDeadlineWins(t *testing.T) {
	calls := 0
	secondRemaining := time.Duration(0)
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		deadline, ok := r.Context().Deadline()
		if !ok {
			return nil, errors.New("missing deadline")
		}
		remaining := time.Until(deadline)
		if calls == 1 {
			if remaining > 180*time.Millisecond {
				t.Errorf("first stage exceeds total budget: %v", remaining)
			}
			return response(r, http.StatusOK, io.NopCloser(strings.NewReader("first"))), nil
		}
		secondRemaining = remaining
		<-r.Context().Done()
		return nil, r.Context().Err()
	})}
	started := time.Now()
	got, err := FetchAll(context.Background(), client, []string{"http://example.test/first", "http://example.test/second"}, 60*time.Millisecond, time.Second)
	if !errors.Is(err, context.DeadlineExceeded) || len(got) != 1 || calls != 2 || time.Since(started) > 500*time.Millisecond {
		t.Fatalf("bodies=%q err=%v calls=%d elapsed=%v", got, err, calls, time.Since(started))
	}
	if secondRemaining <= 0 || secondRemaining > 60*time.Millisecond {
		t.Fatalf("second stage remaining total budget=%v", secondRemaining)
	}
}

func TestFailureRetainsPrefixAndClosesEveryResponseBody(t *testing.T) {
	readErr := errors.New("read failed")
	closeErr := errors.New("close failed")
	first := &trackingBody{reader: strings.NewReader("accepted")}
	second := &trackingBody{readErr: readErr, closeErr: closeErr}
	calls := 0
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return response(r, http.StatusOK, first), nil
		}
		return response(r, http.StatusOK, second), nil
	})}
	got, err := FetchAll(context.Background(), client, []string{"http://example.test/first", "http://example.test/second", "http://example.test/third"}, time.Second, time.Second)
	if !errors.Is(err, readErr) || !errors.Is(err, closeErr) || len(got) != 1 || string(got[0]) != "accepted" {
		t.Fatalf("bodies=%q err=%v", got, err)
	}
	if !first.closed || !second.closed || calls != 2 {
		t.Fatalf("closed=(%v,%v) calls=%d", first.closed, second.closed, calls)
	}
}

func TestStatusAndCloseFailuresRemainInspectable(t *testing.T) {
	statusCloseErr := errors.New("status body close failed")
	body := &trackingBody{reader: strings.NewReader("ignored"), closeErr: statusCloseErr}
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		return response(r, http.StatusPartialContent, body), nil
	})}
	got, err := FetchAll(context.Background(), client, []string{"http://example.test/"}, time.Second, time.Second)
	if err == nil || !strings.Contains(err.Error(), "status 206") || !errors.Is(err, statusCloseErr) || len(got) != 0 || !body.closed {
		t.Fatalf("bodies=%q err=%v closed=%v", got, err, body.closed)
	}
}

func TestFailurePreservesIndependentErrorAndCustomCancellationCause(t *testing.T) {
	operationErr := errors.New("transport failure")
	cause := errors.New("caller stopped for a custom reason")
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		cancel(cause)
		return nil, operationErr
	})}
	got, err := FetchAll(ctx, client, []string{"http://example.test/"}, time.Second, time.Second)
	if !errors.Is(err, operationErr) || !errors.Is(err, context.Canceled) || !errors.Is(err, cause) || len(got) != 0 {
		t.Fatalf("bodies=%q err=%v", got, err)
	}
}

func TestEmptyEndpointsSucceedWithoutCalls(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("unexpected request")
	})}
	got, err := FetchAll(context.Background(), client, nil, time.Second, time.Second)
	if err != nil || len(got) != 0 || calls != 0 {
		t.Fatalf("bodies=%q err=%v calls=%d", got, err, calls)
	}
}
