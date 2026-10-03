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

// Added after inspecting both development outputs. This is a diagnostic for
// dependency ownership, not a frozen comparison assertion or uplift measure.
type recoveryRedirectBody struct {
	io.Reader
	closes    int
	repeatErr error
}

func (b *recoveryRedirectBody) Close() error {
	b.closes++
	if b.closes > 1 {
		return b.repeatErr
	}
	return nil
}

type recoveryRedirectTransport func(*http.Request) (*http.Response, error)

func (f recoveryRedirectTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestRecoveryRedirectBodyOwnership(t *testing.T) {
	policyErr := errors.New("redirect policy rejected target")
	repeatErr := errors.New("body closed twice")
	body := &recoveryRedirectBody{Reader: strings.NewReader("redirect"), repeatErr: repeatErr}
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return policyErr },
		Transport: recoveryRedirectTransport(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusFound,
				Header:     http.Header{"Location": []string{"http://example.test/next"}},
				Body:       body,
				Request:    r,
			}, nil
		}),
	}
	got, err := FetchAll(context.Background(), client, []string{"http://example.test/one"}, time.Second, time.Second)
	if len(got) != 0 || body.closes != 1 || !errors.Is(err, policyErr) || errors.Is(err, repeatErr) {
		t.Fatalf("bodies=%q closes=%d policy=%v repeated-close-error=%v error=%v", got, body.closes, errors.Is(err, policyErr), errors.Is(err, repeatErr), err)
	}
}
