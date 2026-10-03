package stages

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

type recoveryBody struct {
	read  func([]byte) (int, error)
	close func() error
}

func (b recoveryBody) Read(p []byte) (int, error) { return b.read(p) }
func (b recoveryBody) Close() error               { return b.close() }

func TestRecoveryEmptyAndCompletedSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := &http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("empty operation issued request")
		return nil, nil
	})}
	if got, err := FetchAll(ctx, client, nil, time.Second, time.Second); err != nil || len(got) != 0 {
		t.Fatalf("empty result=%q error=%v", got, err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Request: r, Body: recoveryBody{
			read: func([]byte) (int, error) { return 0, io.EOF },
			close: func() error {
				cancel()
				return nil
			},
		}}, nil
	})
	if got, err := FetchAll(ctx, client, []string{"http://example.test/complete"}, time.Second, time.Second); err != nil || len(got) != 1 {
		t.Fatalf("completed result=%q error=%v", got, err)
	}
}

func TestRecoveryBudgetAndBodyLifetime(t *testing.T) {
	for _, shorterParent := range []bool{false, true} {
		ctx := context.Background()
		if shorterParent {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Second)
			defer cancel()
		}
		var scopes []context.Context
		client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
			if len(scopes) > 0 && scopes[0].Err() == nil {
				t.Error("previous stage cancel resources still live")
			}
			scopes = append(scopes, r.Context())
			deadline, ok := r.Context().Deadline()
			if !ok {
				t.Error("request has no total deadline")
			} else if len(scopes) > 1 {
				first, _ := scopes[0].Deadline()
				if !deadline.Equal(first) {
					t.Error("total deadline reset across stages")
				}
			}
			if parent, ok := ctx.Deadline(); ok && !deadline.Equal(parent) {
				t.Error("earlier parent deadline not retained")
			}
			checkLive := func() {
				if r.Context().Err() != nil {
					t.Error("stage canceled before body read/close completed")
				}
			}
			return &http.Response{StatusCode: 200, Header: make(http.Header), Request: r, Body: recoveryBody{
				read:  func([]byte) (int, error) { checkLive(); return 0, io.EOF },
				close: func() error { checkLive(); return nil },
			}}, nil
		})}
		got, err := FetchAll(ctx, client, []string{"http://example.test/one", "http://example.test/two"}, 2*time.Second, 10*time.Second)
		if err != nil || len(got) != 2 {
			t.Fatalf("result=%q error=%v", got, err)
		}
		for _, scope := range scopes {
			if scope.Err() == nil {
				t.Error("stage resources retained after return")
			}
		}
	}
	start := time.Now()
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		deadline, ok := r.Context().Deadline()
		if !ok || deadline.After(start.Add(2100*time.Millisecond)) {
			t.Error("stage ceiling not enforced")
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Request: r, Body: recoveryBody{
			read: func([]byte) (int, error) { return 0, io.EOF }, close: func() error { return nil },
		}}, nil
	})}
	if _, err := FetchAll(context.Background(), client, []string{"http://example.test/one"}, 10*time.Second, 2*time.Second); err != nil {
		t.Fatal(err)
	}
}

type recoveryCause struct{ reasons []string }

func (recoveryCause) Error() string { return "caller stopped" }

func TestRecoveryIndependentFailuresAndPrefix(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	readErr, closeErr := errors.New("read fault"), errors.New("close fault")
	calls, closes := 0, 0
	client := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls > 2 {
			t.Fatal("later request admitted after failure")
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Request: r, Body: recoveryBody{
			read: func([]byte) (int, error) {
				if calls == 1 {
					return 0, io.EOF
				}
				cancel(recoveryCause{[]string{"maintenance"}})
				return 0, errors.Join(context.Canceled, readErr)
			},
			close: func() error {
				closes++
				if calls == 2 {
					return closeErr
				}
				return nil
			},
		}}, nil
	})}
	got, err := FetchAll(ctx, client, []string{"http://example.test/one", "http://example.test/two", "http://example.test/three"}, time.Second, time.Second)
	var cause recoveryCause
	if len(got) != 1 || calls != 2 || closes != 2 || !errors.Is(err, context.Canceled) || !errors.Is(err, readErr) || !errors.Is(err, closeErr) || !errors.As(err, &cause) {
		t.Fatalf("prefix=%q calls=%d closes=%d error=%v", got, calls, closes, err)
	}
}
