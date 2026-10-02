package metadata_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"example.com/metadata"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// These doubles exercise Fetch's supplied-client boundary and body ownership.
// They do not establish network behavior; the server tests below do that.
type observedBody struct {
	io.Reader
	closeCalls int
}

func (b *observedBody) Close() error {
	b.closeCalls++
	return nil
}

func TestFetchSuccess(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		data string
		want metadata.Metadata
	}{
		{"ordinary", `{"name":"first","revision":1}`, metadata.Metadata{Name: "first", Revision: 1}},
		{"preserved_name", `{"name":" \tfirst\n ","revision":7}`, metadata.Metadata{Name: " \tfirst\n ", Revision: 7}},
		{"surrounding_whitespace", " \n {\"name\":\"café\",\"revision\":2} \t\n", metadata.Metadata{Name: "café", Revision: 2}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "caller")
			const endpoint = "https://example.invalid/catalog/a%2Fb?revision=7&name=a%20b"
			body := &observedBody{Reader: strings.NewReader(test.data)}
			var request *http.Request
			calls := 0
			client := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
				request = r
				calls++
				return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
			})}
			got, err := metadata.Fetch(ctx, client, endpoint)
			if body.closeCalls == 0 {
				t.Error("Fetch did not close its response body")
			}
			if err != nil {
				t.Fatalf("Fetch: %v", err)
			}
			if got != test.want {
				t.Errorf("Fetch = %#v, want %#v", got, test.want)
			}
			if calls != 1 || request == nil {
				t.Fatalf("transport calls = %d, want 1", calls)
			}
			if request.Method != http.MethodGet || request.URL.String() != endpoint || request.Body != nil {
				t.Errorf("request = %s %s, body %v; want GET %s with no body", request.Method, request.URL, request.Body, endpoint)
			}
			if request.Context().Value(contextKey{}) != "caller" {
				t.Error("request lost caller context value")
			}
		})
	}
}

func TestFetchResponseErrors(t *testing.T) {
	t.Parallel()
	const valid = `{"name":"first","revision":1}`
	for _, test := range []struct {
		name   string
		status int
		data   string
	}{
		{"created", http.StatusCreated, valid},
		{"no_content", http.StatusNoContent, valid},
		{"redirect", http.StatusFound, valid},
		{"not_found", http.StatusNotFound, valid},
		{"server_error", http.StatusInternalServerError, valid},
		{"empty", http.StatusOK, ""},
		{"whitespace", http.StatusOK, " \t\n"},
		{"malformed", http.StatusOK, `{"name":`},
		{"null", http.StatusOK, `null`},
		{"array", http.StatusOK, `[{"name":"first","revision":1}]`},
		{"scalar", http.StatusOK, `"first"`},
		{"trailing_object", http.StatusOK, valid + ` {"name":"second","revision":2}`},
		{"trailing_null", http.StatusOK, valid + ` null`},
		{"trailing_garbage", http.StatusOK, valid + ` garbage`},
		{"wrong_name_type", http.StatusOK, `{"name":1,"revision":1}`},
		{"wrong_revision_type", http.StatusOK, `{"name":"first","revision":"1"}`},
		{"fractional_revision", http.StatusOK, `{"name":"first","revision":1.5}`},
		{"empty_object", http.StatusOK, `{}`},
		{"missing_name", http.StatusOK, `{"revision":1}`},
		{"empty_name", http.StatusOK, `{"name":"","revision":1}`},
		{"blank_name", http.StatusOK, `{"name":" \t\n\u2003","revision":1}`},
		{"null_name", http.StatusOK, `{"name":null,"revision":1}`},
		{"missing_revision", http.StatusOK, `{"name":"first"}`},
		{"zero_revision", http.StatusOK, `{"name":"first","revision":0}`},
		{"negative_revision", http.StatusOK, `{"name":"first","revision":-1}`},
		{"null_revision", http.StatusOK, `{"name":"first","revision":null}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			body := &observedBody{Reader: strings.NewReader(test.data)}
			client := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: test.status, Body: body}, nil
			})}
			got, err := metadata.Fetch(context.Background(), client, "https://example.invalid/catalog")
			if body.closeCalls == 0 {
				t.Error("Fetch did not close its response body on failure")
			}
			if got != (metadata.Metadata{}) {
				t.Errorf("Fetch = %#v, want zero Metadata on failure", got)
			}
			if err == nil || err.Error() == "" {
				t.Fatalf("Fetch error = %v, want useful error", err)
			}
			if test.status != http.StatusOK && !strings.Contains(err.Error(), strconv.Itoa(test.status)) {
				t.Errorf("Fetch error = %v, want status %d", err, test.status)
			}
		})
	}
}

func TestFetchReadError(t *testing.T) {
	t.Parallel()
	readErr := errors.New("response stream failed")
	// A complete valid object before the read error must not yield partial success.
	body := &observedBody{Reader: io.MultiReader(strings.NewReader(`{"name":"first","revision":1}`), iotest.ErrReader(readErr))}
	client := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
	})}
	got, err := metadata.Fetch(context.Background(), client, "https://example.invalid/catalog")
	if got != (metadata.Metadata{}) || !errors.Is(err, readErr) {
		t.Errorf("Fetch = %#v, %v; want zero Metadata and read error", got, err)
	}
	if body.closeCalls == 0 {
		t.Error("Fetch did not close its response body after a read error")
	}
}

func TestFetchTransportError(t *testing.T) {
	t.Parallel()
	transportErr := errors.New("transport failed")
	calls := 0
	client := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, transportErr
	})}
	got, err := metadata.Fetch(context.Background(), client, "https://example.invalid/catalog")
	if got != (metadata.Metadata{}) || !errors.Is(err, transportErr) {
		t.Errorf("Fetch = %#v, %v; want zero Metadata and transport error", got, err)
	}
	if calls != 1 {
		t.Errorf("transport calls = %d, want 1", calls)
	}
}

func TestFetchInvalidEndpoint(t *testing.T) {
	t.Parallel()
	calls := 0
	client := &http.Client{Transport: roundTrip(func(*http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("unexpected transport call")
	})}
	got, err := metadata.Fetch(context.Background(), client, "http://example.invalid/%")
	if got != (metadata.Metadata{}) || err == nil || err.Error() == "" {
		t.Errorf("Fetch = %#v, %v; want zero Metadata and URL error", got, err)
	}
	if calls != 0 {
		t.Errorf("transport calls = %d, want 0 for an invalid URL", calls)
	}
}

func TestFetchHTTP(t *testing.T) {
	t.Parallel()
	type requestInfo struct{ method, uri string }
	requests := make(chan requestInfo, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case requests <- requestInfo{r.Method, r.RequestURI}:
		default:
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"name":" remote ","revision":3}`)
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	client.Timeout = 5 * time.Second
	t.Cleanup(client.CloseIdleConnections)
	const pathQuery = "/catalog/a%2Fb?revision=7&name=a%20b"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	got, err := metadata.Fetch(ctx, client, server.URL+pathQuery)
	if err != nil {
		t.Fatalf("Fetch through local HTTP: %v", err)
	}
	if want := (metadata.Metadata{Name: " remote ", Revision: 3}); got != want {
		t.Errorf("Fetch = %#v, want %#v", got, want)
	}
	request := receive(t, requests, "server request")
	if request.method != http.MethodGet || request.uri != pathQuery {
		t.Errorf("server received %s %s, want GET %s", request.method, request.uri, pathQuery)
	}
}

type fetchResult struct {
	metadata metadata.Metadata
	err      error
}

// A read signal establishes that Fetch reached the real response stream.
type signaledBody struct {
	io.ReadCloser
	started    chan struct{}
	closeCalls int
}

func (b *signaledBody) Read(p []byte) (int, error) {
	if b.started != nil {
		close(b.started)
		b.started = nil
	}
	return b.ReadCloser.Read(p)
}

func (b *signaledBody) Close() error {
	b.closeCalls++
	return b.ReadCloser.Close()
}

func TestFetchCancellation(t *testing.T) {
	t.Parallel()
	for _, stage := range []string{"before_headers", "during_body"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			started := make(chan error, 1)
			handlerDone := make(chan error, 1)
			release := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var err error
				if stage == "during_body" {
					_, err = io.WriteString(w, `{"name":"`)
					if err == nil {
						err = http.NewResponseController(w).Flush()
					}
				}
				select {
				case started <- err:
				default:
				}
				select {
				case <-r.Context().Done():
				case <-release:
				}
				select {
				case handlerDone <- r.Context().Err():
				default:
				}
			}))
			t.Cleanup(server.Close)
			client := server.Client()
			client.Timeout = 5 * time.Second
			t.Cleanup(client.CloseIdleConnections)
			readStarted := make(chan struct{})
			var body *signaledBody
			if stage == "during_body" {
				transport := client.Transport
				client.Transport = roundTrip(func(r *http.Request) (*http.Response, error) {
					response, err := transport.RoundTrip(r)
					if err != nil {
						return response, err
					}
					body = &signaledBody{ReadCloser: response.Body, started: readStarted}
					response.Body = body
					return response, nil
				})
				// The wrapper forwards real I/O and cancellation; own the underlying pool.
				t.Cleanup(transport.(*http.Transport).CloseIdleConnections)
			}
			ctx, cancel := context.WithCancel(context.Background())
			results := make(chan fetchResult, 1)
			fetchDone := make(chan struct{})
			// Run on early failures too, releasing handlers before joining the worker
			// and before server.Close waits for handlers to finish.
			t.Cleanup(func() {
				cancel()
				close(release)
				timer := time.NewTimer(5 * time.Second)
				defer timer.Stop()
				select {
				case <-fetchDone:
				case <-timer.C:
					t.Error("Fetch worker did not finish during cleanup")
				}
			})
			go func() {
				defer close(fetchDone)
				got, err := metadata.Fetch(ctx, client, server.URL+"/catalog")
				results <- fetchResult{got, err}
			}()
			if err := receive(t, started, "handler startup"); err != nil {
				t.Fatalf("start response stream: %v", err)
			}
			if stage == "during_body" {
				receive(t, readStarted, "response body read")
			}
			cancel()
			result := receive(t, results, "Fetch completion after caller cancellation")
			if result.metadata != (metadata.Metadata{}) || !errors.Is(result.err, context.Canceled) {
				t.Errorf("Fetch = %#v, %v; want zero Metadata and context.Canceled", result.metadata, result.err)
			}
			if err := receive(t, handlerDone, "server cancellation"); !errors.Is(err, context.Canceled) {
				t.Errorf("server context error = %v, want context.Canceled", err)
			}
			if stage == "during_body" && (body == nil || body.closeCalls == 0) {
				t.Error("Fetch did not close its acquired response body after cancellation")
			}
		})
	}
}

// This deadline diagnoses stalled work, rather than guessing event readiness.
func receive[T any](t *testing.T, events <-chan T, operation string) T {
	t.Helper()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case event := <-events:
		return event
	case <-timer.C:
		t.Fatalf("timed out waiting for %s", operation)
		var zero T
		return zero
	}
}
