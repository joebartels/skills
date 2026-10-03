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

type reviewBody struct { reader io.Reader; close func() error }
func (b *reviewBody) Read(p []byte) (int,error) { return b.reader.Read(p) }
func (b *reviewBody) Close() error { return b.close() }

func TestReviewStageScopeThroughClose(t *testing.T) {
    closes := 0
    client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response,error) {
        b := &reviewBody{reader:strings.NewReader("complete"),close:func() error {
            closes++
            <-r.Context().Done()
            return r.Context().Err()
        }}
        return response(r,200,b),nil
    })}
    got,err := FetchAll(context.Background(),client,[]string{"http://example.test/a"},time.Second,25*time.Millisecond)
    if len(got)!=0 || !errors.Is(err,context.DeadlineExceeded) || closes!=1 { t.Fatalf("bodies=%q err=%v closes=%d",got,err,closes) }
}

func TestReviewReleaseBeforeNextStage(t *testing.T) {
    var previous context.Context
    calls := 0
    client := &http.Client{Transport:transportFunc(func(r *http.Request)(*http.Response,error) {
        calls++
        if previous!=nil && previous.Err()==nil { return nil,errors.New("previous stage still live") }
        previous=r.Context()
        b:=&reviewBody{reader:strings.NewReader("ok"),close:func()error {
            if r.Context().Err()!=nil { return errors.New("scope canceled before close") }; return nil
        }}
        return response(r,200,b),nil
    })}
    got,err:=FetchAll(context.Background(),client,[]string{"http://example.test/a","http://example.test/b"},time.Second,time.Second)
    if err!=nil || len(got)!=2 || calls!=2 || previous.Err()==nil { t.Fatalf("bodies=%q err=%v calls=%d context=%v",got,err,calls,previous.Err()) }
}

func TestReviewDualReadCloseErrors(t *testing.T) {
    readErr,closeErr:=errors.New("independent read"),errors.New("independent close")
    client:=&http.Client{Transport:transportFunc(func(r *http.Request)(*http.Response,error) {
        return response(r,200,&reviewBody{reader:errorReader{readErr},close:func()error{return closeErr}}),nil
    })}
    got,err:=FetchAll(context.Background(),client,[]string{"http://example.test/a"},time.Second,time.Second)
    if len(got)!=0 || !errors.Is(err,readErr) || !errors.Is(err,closeErr) { t.Fatalf("bodies=%q err=%v",got,err) }
}

func TestReviewCompletedSuccessSurvivesCloseCancellation(t *testing.T) {
    ctx,cancel:=context.WithCancel(context.Background()); defer cancel()
    client:=&http.Client{Transport:transportFunc(func(r *http.Request)(*http.Response,error) {
        return response(r,200,&reviewBody{reader:strings.NewReader("complete"),close:func()error{cancel();return nil}}),nil
    })}
    got,err:=FetchAll(ctx,client,[]string{"http://example.test/a"},time.Second,time.Second)
    if err!=nil || len(got)!=1 || string(got[0])!="complete" { t.Fatalf("bodies=%q err=%v",got,err) }
}

func TestReviewStageDeadlineCeiling(t *testing.T) {
    budget:=50*time.Millisecond
    client:=&http.Client{Transport:transportFunc(func(r *http.Request)(*http.Response,error) {
        deadline,ok:=r.Context().Deadline()
        if !ok || deadline.After(time.Now().Add(budget)) { return nil,errors.New("request deadline exceeds configured stage duration") }
        return response(r,200,io.NopCloser(strings.NewReader("ok"))),nil
    })}
    got,err:=FetchAll(context.Background(),client,[]string{"http://example.test/a"},5*time.Second,budget)
    if err!=nil || len(got)!=1 { t.Fatalf("bodies=%q err=%v",got,err) }
}
