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

type testBody struct {
	reader io.Reader
	close  func() error
	closed atomic.Int32
}

func (b *testBody) Read(p []byte) (int, error) { return b.reader.Read(p) }
func (b *testBody) Close() error {
	b.closed.Add(1)
	if b.close != nil {
		return b.close()
	}
	return nil
}

func response(r *http.Request, status int, body io.ReadCloser) *http.Response {
	return &http.Response{StatusCode: status, Body: body, Header: make(http.Header), Request: r}
}

func TestSequentialBodiesAndStageLifetime(t *testing.T) {
	var calls int
	var firstCtx context.Context
	var closeSawLiveContext bool
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			firstCtx = r.Context()
			body := &testBody{reader: strings.NewReader(r.URL.Path), close: func() error {
				closeSawLiveContext = r.Context().Err() == nil
				return nil
			}}
			return response(r, http.StatusOK, body), nil
		}
		return response(r, http.StatusOK, io.NopCloser(strings.NewReader(r.URL.Path))), nil
	})}
	got, err := FetchAll(context.Background(), client, []string{"http://example.test/first", "http://example.test/second"}, time.Second, time.Second)
	if err != nil || len(got) != 2 || string(got[0]) != "/first" || string(got[1]) != "/second" {
		t.Fatalf("bodies=%q err=%v", got, err)
	}
	if calls != 2 || !closeSawLiveContext {
		t.Fatalf("calls=%d close saw live context=%v", calls, closeSawLiveContext)
	}
	if firstCtx.Err() == nil {
		t.Fatal("stage context remained live after FetchAll returned")
	}
}

func TestFailuresRetainPrefixAndCloseBodies(t *testing.T) {
	readErr := errors.New("read failed")
	closeErr := errors.New("close failed")
	for _, tc := range []struct {
		name   string
		status int
		reader io.Reader
		closer error
		want   []string
		cause  error
	}{
		{name: "non-200 status", status: http.StatusPartialContent, reader: strings.NewReader("ignored"), want: []string{"/good"}},
		{name: "read failure", status: http.StatusOK, reader: errorReader{readErr}, want: []string{"/good"}, cause: readErr},
		{name: "close failure", status: http.StatusOK, reader: strings.NewReader("bad"), closer: closeErr, want: []string{"/good"}, cause: closeErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			var failedBody *testBody
			client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return response(r, http.StatusOK, io.NopCloser(strings.NewReader("/good"))), nil
				}
				failedBody = &testBody{reader: tc.reader, close: func() error { return tc.closer }}
				return response(r, tc.status, failedBody), nil
			})}
			got, err := FetchAll(context.Background(), client, []string{"http://example.test/good", "http://example.test/bad", "http://example.test/later"}, time.Second, time.Second)
			if err == nil || calls != 2 || len(got) != 1 || string(got[0]) != "/good" {
				t.Fatalf("bodies=%q calls=%d err=%v", got, calls, err)
			}
			if tc.cause != nil && !errors.Is(err, tc.cause) {
				t.Errorf("error %v does not include %v", err, tc.cause)
			}
			if failedBody.closed.Load() != 1 {
				t.Errorf("failed body closed %d times", failedBody.closed.Load())
			}
		})
	}
}

type errorReader struct{ err error }

func (r errorReader) Read([]byte) (int, error) { return 0, r.err }

func TestTotalBudgetIsSharedAcrossStages(t *testing.T) {
	var deadlines []time.Time
	missingDeadline := false
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		deadline, ok := r.Context().Deadline()
		if !ok {
			missingDeadline = true
		} else {
			deadlines = append(deadlines, deadline)
		}
		return response(r, http.StatusOK, io.NopCloser(strings.NewReader("ok"))), nil
	})}
	parentDeadline := time.Now().Add(2 * time.Second)
	parent, cancel := context.WithDeadline(context.Background(), parentDeadline)
	defer cancel()
	_, err := FetchAll(parent, client, []string{"http://example.test/a", "http://example.test/b"}, time.Second, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if missingDeadline || len(deadlines) != 2 || !deadlines[0].Equal(deadlines[1]) {
		t.Fatalf("request deadlines = %v, want one shared total deadline", deadlines)
	}
	if deadlines[0].After(parentDeadline) {
		t.Fatalf("request deadline %v exceeds parent %v", deadlines[0], parentDeadline)
	}
}

func TestAlreadyCanceledCallStartsNoRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("unexpected request")
	})}
	got, err := FetchAll(ctx, client, nil, time.Second, time.Second)
	if !errors.Is(err, context.Canceled) || calls != 0 || len(got) != 0 {
		t.Fatalf("bodies=%v calls=%d err=%v", got, calls, err)
	}
}

func TestCancellationAtReadFailurePreservesCauseAndReadError(t *testing.T) {
	cause := testCause{marker: []int{1}}
	readErr := errors.New("reader stopped")
	ctx, cancel := context.WithCancelCause(context.Background())
	started := make(chan struct{})
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		body := &testBody{reader: cancelReader{ctx: r.Context(), started: started, err: readErr}}
		return response(r, http.StatusOK, body), nil
	})}
	done := make(chan error, 1)
	go func() {
		_, err := FetchAll(ctx, client, []string{"http://example.test/a"}, time.Second, time.Second)
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("body read did not start")
	}
	cancel(cause)
	select {
	case err := <-done:
		var gotCause testCause
		if !errors.Is(err, readErr) || !errors.Is(err, context.Canceled) || !errors.As(err, &gotCause) || len(gotCause.marker) != 1 || gotCause.marker[0] != 1 {
			t.Fatalf("error %v lacks independent read/cancellation failures", err)
		}
	case <-time.After(time.Second):
		t.Fatal("FetchAll did not return after cooperative body cancellation")
	}
}

type testCause struct{ marker []int }

func (testCause) Error() string { return "custom cancellation cause" }

type cancelReader struct {
	ctx     context.Context
	started chan<- struct{}
	err     error
}

func (r cancelReader) Read([]byte) (int, error) {
	r.started <- struct{}{}
	<-r.ctx.Done()
	return 0, r.err
}
