package stages

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestTotalAndEarlierParentDeadline(t *testing.T) {
	for _, earlier := range []bool{false, true} {
		t.Run(map[bool]string{false: "total", true: "parent"}[earlier], func(t *testing.T) {
			ctx := context.Background()
			parentDeadline := time.Now().Add(5 * time.Second)
			if earlier {
				var cancel context.CancelFunc
				ctx, cancel = context.WithDeadline(ctx, parentDeadline)
				defer cancel()
			}
			start := time.Now()
			var deadlines []time.Time
			client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				d, ok := r.Context().Deadline()
				if !ok {
					return nil, errors.New("missing operation deadline")
				}
				deadlines = append(deadlines, d)
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("complete")), Header: make(http.Header), Request: r}, nil
			})}
			bodies, err := FetchAll(ctx, client, []string{"http://example.test/one", "http://example.test/two"}, 10*time.Second, time.Minute)
			if err != nil || len(bodies) != 2 || len(deadlines) != 2 {
				t.Fatalf("bodies=%q deadlines=%v err=%v", bodies, deadlines, err)
			}
			if !deadlines[0].Equal(deadlines[1]) || deadlines[0].After(start.Add(10*time.Second+10*time.Millisecond)) {
				t.Fatalf("total scope reset: %v", deadlines)
			}
			if earlier && !deadlines[0].Equal(parentDeadline) {
				t.Fatalf("earlier parent lost: got=%v want=%v", deadlines[0], parentDeadline)
			}
		})
	}
}

type heldBody struct {
	ctx     context.Context
	started chan struct{}
	closed  atomic.Bool
}

func (b *heldBody) Read([]byte) (int, error) { close(b.started); <-b.ctx.Done(); return 0, b.ctx.Err() }
func (b *heldBody) Close() error             { b.closed.Store(true); return nil }

func TestResponseScopeAndClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var body *heldBody
	readStarted := make(chan struct{})
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/first" {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("kept")), Header: make(http.Header), Request: r}, nil
		}
		body = &heldBody{ctx: r.Context(), started: readStarted}
		return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header), Request: r}, nil
	})}
	type outcome struct {
		bodies [][]byte
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		got, err := FetchAll(ctx, client, []string{"http://example.test/first", "http://example.test/held"}, time.Second, time.Second)
		done <- outcome{got, err}
	}()
	select {
	case <-readStarted:
	case got := <-done:
		t.Fatalf("premature response completion: %+v", got)
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("body read did not start")
	}
	if body.closed.Load() {
		cancel()
		t.Fatal("body closed while read is active")
	}
	cancel()
	select {
	case got := <-done:
		if len(got.bodies) != 1 || string(got.bodies[0]) != "kept" || !errors.Is(got.err, context.Canceled) || !body.closed.Load() {
			t.Fatalf("bodies=%q err=%v closed=%v", got.bodies, got.err, body.closed.Load())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("canceled body read did not finish")
	}
}

func TestStageBudgetAndCanceledAdmission(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) { calls++; return nil, r.Context().Err() })}
	_, err := FetchAll(ctx, client, []string{"http://example.test/one"}, time.Second, time.Second)
	if calls != 0 || !errors.Is(err, context.Canceled) {
		t.Fatalf("already canceled: calls=%d err=%v", calls, err)
	}
	readStarted := make(chan struct{})
	var body *heldBody
	client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		body = &heldBody{ctx: r.Context(), started: readStarted}
		return &http.Response{StatusCode: 200, Body: body, Header: make(http.Header), Request: r}, nil
	})
	type outcome struct{ err error }
	done := make(chan outcome, 1)
	parent, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()
	go func() {
		_, err := FetchAll(parent, client, []string{"http://example.test/held"}, time.Second, 30*time.Millisecond)
		done <- outcome{err}
	}()
	select {
	case <-readStarted:
	case <-time.After(2 * time.Second):
		parentCancel()
		t.Fatal("stage did not start")
	}
	select {
	case got := <-done:
		if !errors.Is(got.err, context.DeadlineExceeded) || !body.closed.Load() {
			t.Fatalf("stage err=%v closed=%v", got.err, body.closed.Load())
		}
	case <-time.After(2 * time.Second):
		parentCancel()
		<-done
		t.Fatal("stage budget did not stop body consumption")
	}
}

func TestHTTPCancellationBoundary(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("local listener unavailable: %v", err)
	}
	started, finished := make(chan struct{}), make(chan struct{})
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/first" {
			io.WriteString(w, "retained")
			return
		}
		w.WriteHeader(200)
		io.WriteString(w, "partial")
		w.(http.Flusher).Flush()
		close(started)
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
		close(finished)
	}))
	server.Listener.Close()
	server.Listener = listener
	server.Start()
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	type outcome struct {
		bodies [][]byte
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		b, e := FetchAll(ctx, server.Client(), []string{server.URL + "/first", server.URL + "/held"}, time.Second, time.Second)
		done <- outcome{b, e}
	}()
	select {
	case <-started:
	case got := <-done:
		t.Fatalf("premature HTTP return: %+v", got)
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP stage not started")
	}
	cancel()
	select {
	case got := <-done:
		if len(got.bodies) != 1 || string(got.bodies[0]) != "retained" || !errors.Is(got.err, context.Canceled) {
			t.Fatalf("HTTP bodies=%q err=%v", got.bodies, got.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP cancellation did not return")
	}
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("HTTP handler did not observe cancellation")
	}
}
