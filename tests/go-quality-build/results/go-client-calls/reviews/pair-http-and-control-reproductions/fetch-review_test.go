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

type reviewRT func(*http.Request) (*http.Response, error)

func (f reviewRT) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type reviewReader func([]byte) (int, error)

func (f reviewReader) Read(p []byte) (int, error) { return f(p) }
func reviewResponse(status int, body io.ReadCloser) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: body}
}

func TestReviewFetchRedirectTerminal(t *testing.T) {
	calls := 0
	c := &http.Client{Transport: reviewRT(func(r *http.Request) (*http.Response, error) {
		calls++
		status := 200
		if calls == 1 {
			status = 302
		}
		resp := reviewResponse(status, io.NopCloser(strings.NewReader("result")))
		if status == 302 {
			resp.Header.Set("Location", "/next")
		}
		return resp, nil
	})}
	got, err := Fetch(context.Background(), c, "http://example.invalid/start")
	t.Logf("redirect: calls=%d result=%q error=%v", calls, got, err)
	if calls != 1 || err == nil || !strings.Contains(err.Error(), "HTTP 302") {
		t.Error("302 was not terminal")
	}
}

func TestReviewFetchPolicyStatus(t *testing.T) {
	cause := errors.New("caller rejects redirect")
	c := &http.Client{Transport: reviewRT(func(*http.Request) (*http.Response, error) {
		resp := reviewResponse(302, io.NopCloser(strings.NewReader("redirect")))
		resp.Header.Set("Location", "/next")
		return resp, nil
	}), CheckRedirect: func(*http.Request, []*http.Request) error { return cause }}
	_, err := Fetch(context.Background(), c, "http://example.invalid/start")
	t.Logf("redirect policy: cause=%t error=%v", errors.Is(err, cause), err)
	if !errors.Is(err, cause) || !strings.Contains(err.Error(), "HTTP 302") {
		t.Error("status or policy cause lost")
	}
}

func TestReviewFetchLongDiagnostics(t *testing.T) {
	cause := errors.New(strings.Repeat("e", 10000))
	c := &http.Client{Transport: reviewRT(func(*http.Request) (*http.Response, error) { return nil, cause })}
	_, err := Fetch(context.Background(), c, "http://example.invalid")
	t.Logf("diagnostic bytes=%d cause=%t", len(err.Error()), errors.Is(err, cause))
	if len(err.Error()) > 4096 || !errors.Is(err, cause) {
		t.Error("diagnostic not capped independently of error identity")
	}
}

func TestReviewFetchClientTimeoutCauses(t *testing.T) {
	cause := errors.New("independent read failure")
	closes := 0
	c := &http.Client{Timeout: 15 * time.Millisecond, Transport: reviewRT(func(r *http.Request) (*http.Response, error) {
		return reviewResponse(200, &reviewCloseBody{reader: reviewReader(func([]byte) (int, error) { <-r.Context().Done(); return 0, errors.Join(cause, r.Context().Err()) }), closes: &closes}), nil
	})}
	_, err := Fetch(context.Background(), c, "http://example.invalid")
	t.Logf("client timeout: bodyCause=%t deadline=%t closes=%d error=%v", errors.Is(err, cause), errors.Is(err, context.DeadlineExceeded), closes, err)
	if !errors.Is(err, cause) || !errors.Is(err, context.DeadlineExceeded) || closes != 1 {
		t.Error("client timeout masked underlying failures")
	}
}

type reviewCloseBody struct {
	reader io.Reader
	closes *int
}

func (b *reviewCloseBody) Read(p []byte) (int, error) { return b.reader.Read(p) }
func (b *reviewCloseBody) Close() error               { *b.closes++; return nil }
