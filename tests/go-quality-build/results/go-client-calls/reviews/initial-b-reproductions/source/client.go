package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	submitBudget = 500 * time.Millisecond
	retryBackoff = 10 * time.Millisecond
	maxBodyBytes = 4096
)

// Submit posts payload within 500ms, accepting only HTTP 200.
// With server deduplication, operationID must be nonempty and identify the same
// payload across invocations. Transport failures can leave the outcome unknown.
// The supplied client and its transport remain owned by the caller.
func Submit(ctx context.Context, client *http.Client, endpoint, operationID string, payload []byte, deduplicates bool) ([]byte, error) {
	if deduplicates && operationID == "" {
		return nil, errors.New("deduplicated POST requires an operation identity")
	}
	ctx, cancel := context.WithTimeout(ctx, submitBudget)
	defer cancel()
	payload = bytes.Clone(payload)

	// Do must expose the initial status even if the client's policy permits a
	// redirect. Keep caller policy errors and never mutate the borrowed client.
	attemptClient := *client
	attemptClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if client.CheckRedirect != nil {
			if err := client.CheckRedirect(req, via); err != nil {
				return err
			}
		}
		return http.ErrUseLastResponse
	}

	attempts := 1
	if deduplicates {
		attempts = 3
	}
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, errors.Join(lastErr, err, context.Cause(ctx))
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		if deduplicates {
			req.Header.Set("Idempotency-Key", operationID)
		}
		body, err, retry := submitAttempt(&attemptClient, req)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if stopErr := ctx.Err(); stopErr != nil {
			return nil, errors.Join(err, stopErr, context.Cause(ctx))
		}
		if !retry || attempt+1 == attempts {
			return nil, err
		}
		timer := time.NewTimer(retryBackoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, errors.Join(err, ctx.Err(), context.Cause(ctx))
		case <-timer.C:
			timer.Stop()
		}
	}
	return nil, lastErr
}

func submitAttempt(client *http.Client, req *http.Request) ([]byte, error, bool) {
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	observed := &attemptTransport{RoundTripper: transport}
	attemptClient := *client
	attemptClient.Transport = observed
	resp, err := attemptClient.Do(req)
	if err != nil {
		if observed.err != nil && !errors.Is(err, observed.err) {
			err = errors.Join(err, observed.err)
		}
		// Client policy and malformed redirects can fail after a response.
		// Do has already closed that body, including when resp is nil.
		if observed.status != 0 && observed.err == nil {
			return nil, fmt.Errorf("HTTP %d: %w", observed.status, err), false
		}
		return nil, fmt.Errorf("POST transport failed; outcome unknown: %w", err), true
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	if readErr != nil && observed.bodyErr != nil && !errors.Is(readErr, observed.bodyErr) {
		readErr = errors.Join(readErr, observed.bodyErr)
	}
	if resp.StatusCode != http.StatusOK {
		if len(body) > maxBodyBytes {
			body = body[:maxBodyBytes]
		}
		statusErr := fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
		if readErr != nil {
			statusErr = errors.Join(statusErr, readErr)
		}
		return nil, statusErr, resp.StatusCode == http.StatusServiceUnavailable
	}
	if readErr != nil {
		return nil, fmt.Errorf("read POST response; outcome unknown: %w", readErr), true
	}
	if len(body) > maxBodyBytes {
		return nil, fmt.Errorf("HTTP 200 response exceeds %d bytes", maxBodyBytes), false
	}
	return body, nil, false
}

// attemptTransport distinguishes failures in the transport from client policy.
type attemptTransport struct {
	http.RoundTripper
	status  int
	err     error
	bodyErr error
}

func (t *attemptTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.RoundTripper.RoundTrip(req)
	t.err = err
	if resp != nil {
		t.status = resp.StatusCode
		if err == nil && resp.Body != nil {
			resp.Body = &attemptBody{ReadCloser: resp.Body, transport: t}
		}
	}
	return resp, err
}

func (t *attemptTransport) CancelRequest(req *http.Request) {
	// Keep cancellation support for legacy caller transports.
	if transport, ok := t.RoundTripper.(interface{ CancelRequest(*http.Request) }); ok {
		transport.CancelRequest(req)
	}
}

// attemptBody retains read errors that Client.Timeout may replace.
type attemptBody struct {
	io.ReadCloser
	transport *attemptTransport
}

func (b *attemptBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err != nil && err != io.EOF {
		b.transport.bodyErr = err
	}
	return n, err
}
