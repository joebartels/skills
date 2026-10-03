package stages

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type reviewBody struct {
	io.Reader
	closeFn func() error
}

func (b reviewBody) Close() error { return b.closeFn() }

func TestReviewOneTotalDeadlineAndEarlierParentDeadline(t *testing.T) {
	for _, parentBound := range []bool{false, true} {
		t.Run(map[bool]string{false: "total", true: "parent"}[parentBound], func(t *testing.T) {
			ctx := context.Background()
			var parentDeadline time.Time
			if parentBound {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
				defer cancel()
				parentDeadline, _ = ctx.Deadline()
			}
			var firstDeadline time.Time
			calls := 0
			client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				d, ok := r.Context().Deadline()
				if !ok { t.Error("missing request deadline") }
				if calls == 1 { firstDeadline = d }
				if d != firstDeadline { t.Errorf("total scope deadline restarted: first=%v current=%v", firstDeadline, d) }
				if parentBound && d != parentDeadline { t.Errorf("parent deadline not retained: got=%v want=%v", d, parentDeadline) }
				return response(r, 200, io.NopCloser(strings.NewReader("ok"))), nil
			})}
			got, err := FetchAll(ctx, client, []string{"http://example.test/1", "http://example.test/2"}, time.Minute, 2*time.Minute)
			if err != nil || len(got) != 2 || calls != 2 { t.Fatalf("bodies=%q err=%v calls=%d",got,err,calls) }
		})
	}
}

func TestReviewStageScopeLiveUntilCloseAndReleasedBeforeNextRequest(t *testing.T) {
	var prior context.Context
	calls := 0
	closed := 0
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if prior != nil && prior.Err() == nil { t.Error("prior stage not released before next request") }
		prior = r.Context()
		return response(r, 200, reviewBody{Reader: readerFunc(func(p []byte) (int,error) {
			if r.Context().Err()!=nil { t.Error("stage canceled before body read") }
			return 0,io.EOF
		}),closeFn:func() error {
			if r.Context().Err()!=nil { t.Error("stage canceled before body close") }
			closed++
			return nil
		}}),nil
	})}
	got,err:=FetchAll(context.Background(),client,[]string{"http://example.test/1","http://example.test/2"},time.Second,time.Second)
	if err!=nil || len(got)!=2 || closed!=2 || prior.Err()==nil { t.Fatalf("bodies=%q err=%v closed=%d final-context=%v",got,err,closed,prior.Err()) }
}

type readerFunc func([]byte)(int,error)
func (f readerFunc) Read(p []byte)(int,error) { return f(p) }

func TestReviewCloseOnlyFailureExcludesBody(t *testing.T) {
	want:=errors.New("close-only failure")
	client:=&http.Client{Transport:transportFunc(func(r *http.Request)(*http.Response,error){
		return response(r,200,reviewBody{Reader:strings.NewReader("complete"),closeFn:func()error{return want}}),nil
	})}
	got,err:=FetchAll(context.Background(),client,[]string{"http://example.test/"},time.Second,time.Second)
	if len(got)!=0 || !errors.Is(err,want) { t.Fatalf("bodies=%q err=%v",got,err) }
}

func TestReviewCompleteSuccessNotInvalidatedByCloseCancellation(t *testing.T) {
	ctx,cancel:=context.WithCancel(context.Background())
	defer cancel()
	client:=&http.Client{Transport:transportFunc(func(r *http.Request)(*http.Response,error){
		return response(r,200,reviewBody{Reader:strings.NewReader("complete"),closeFn:func()error{cancel();return nil}}),nil
	})}
	got,err:=FetchAll(ctx,client,[]string{"http://example.test/"},time.Second,time.Second)
	if err!=nil || len(got)!=1 || string(got[0])!="complete" { t.Fatalf("bodies=%q err=%v",got,err) }
}

func TestReviewBodyFailurePreservesCustomCauseAndCloseError(t *testing.T) {
	ctx,cancel:=context.WithCancelCause(context.Background())
	defer cancel(nil)
	cause:=errors.New("custom body cancellation")
	readErr:=errors.New("independent read failure")
	closeErr:=errors.New("independent close failure")
	closed:=false
	client:=&http.Client{Transport:transportFunc(func(r *http.Request)(*http.Response,error){
		return response(r,200,reviewBody{Reader:readerFunc(func([]byte)(int,error){cancel(cause);return 0,readErr}),closeFn:func()error{closed=true;return closeErr}}),nil
	})}
	got,err:=FetchAll(ctx,client,[]string{"http://example.test/"},time.Second,time.Second)
	if len(got)!=0 || !closed || !errors.Is(err,cause) || !errors.Is(err,context.Canceled) || !errors.Is(err,readErr) || !errors.Is(err,closeErr) { t.Fatalf("bodies=%q err=%v closed=%v",got,err,closed) }
}

func TestReviewStandardTransportBodyReadCancellation(t *testing.T) {
	observed:=make(chan struct{})
	server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		close(observed)
	}))
	defer server.Close()
	got,err:=FetchAll(context.Background(),server.Client(),[]string{server.URL},2*time.Second,100*time.Millisecond)
	if len(got)!=0 || !errors.Is(err,context.DeadlineExceeded) { t.Fatalf("bodies=%q err=%v",got,err) }
	select { case <-observed: case <-time.After(time.Second): t.Fatal("local test handler did not observe canceled request") }
}
