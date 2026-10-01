package receipts

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestSend(t *testing.T) {
	t.Setenv("RECEIPT_URL", "https://receipts.invalid/send")
	original := http.DefaultClient
	t.Cleanup(func() { http.DefaultClient = original })
	http.DefaultClient = &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		if r.Method != "POST" || r.URL.String() != "https://receipts.invalid/send" || r.Header.Get("Content-Type") != "application/json" || string(body) != `{"id":"abc"}` {
			t.Errorf("unexpected request: %s %s %q", r.Method, r.URL, body)
		}
		return &http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}
	if err := (Sender{}).Send(context.Background(), "abc"); err != nil {
		t.Fatal(err)
	}
}
