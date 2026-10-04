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

// Submit posts payload with a 500ms total budget and returns a complete HTTP 200
// response of at most 4096 bytes. The caller retains ownership of client.
// With deduplication, operationID must identify the same payload on every retry.
// Without it, a transport failure leaves the operation's effect uncertain.
func Submit(ctx context.Context, client *http.Client, endpoint, operationID string, payload []byte, deduplicates bool) ([]byte, error) {
	if deduplicates && operationID == "" {
		return nil, errors.New("empty operation identity")
	}
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, stopped(ctx, err)
	}
	payload = bytes.Clone(payload)

	local := *client
	redirect := client.CheckRedirect
	local.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if redirect != nil {
			if err := redirect(req, via); err != nil {
				return err
			}
		}
		// A redirect is a terminal status, even when its target would return 200.
		return http.ErrUseLastResponse
	}
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, stopped(ctx, errors.Join(lastErr, err))
		}
		data, err, retry := submitAttempt(ctx, &local, transport, endpoint, operationID, payload, deduplicates)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return nil, stopped(ctx, err)
		}
		if !deduplicates || !retry || attempt == 2 {
			return nil, err
		}
		timer := time.NewTimer(time.Duration(10<<attempt) * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, stopped(ctx, err)
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func stopped(ctx context.Context, err error) error {
	return errors.Join(err, ctx.Err(), context.Cause(ctx))
}

func submitAttempt(ctx context.Context, client *http.Client, transport http.RoundTripper, endpoint, operationID string, payload []byte, deduplicates bool) ([]byte, error, bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err, false
	}
	if deduplicates {
		req.Header.Set("Idempotency-Key", operationID)
	} else {
		req.GetBody = nil
	}
	// Keep legacy transports under the operation scope as well as its context.
	req.Cancel = ctx.Done()
	observed := &observedTransport{next: transport}
	attemptClient := *client
	attemptClient.Transport = observed
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			req.Body.Close()
			return nil, stopped(ctx, context.DeadlineExceeded), false
		}
		if attemptClient.Timeout == 0 || attemptClient.Timeout > remaining {
			attemptClient.Timeout = remaining
		}
	}
	resp, err := attemptClient.Do(req)
	if err != nil {
		// Client.Timeout can replace a transport cause; redirect errors have an
		// already-closed response body owned by Client.Do.
		err = errors.Join(err, observed.transportErr, observed.bodyErr, observed.stopErr)
		if observed.status != 0 && observed.status != http.StatusOK {
			err = errors.Join(fmt.Errorf("HTTP %d", observed.status), err)
			return nil, err, false
		}
		return nil, err, observed.transportErr != nil
	}

	limit := int64(4096)
	if resp.StatusCode == http.StatusOK {
		limit++ // One extra byte distinguishes an oversized complete response.
	}
	data, readErr := io.ReadAll(io.LimitReader(resp.Body, limit))
	closeErr := resp.Body.Close()
	err = errors.Join(readErr, observed.bodyErr, observed.stopErr, closeErr)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Join(fmt.Errorf("HTTP %d", resp.StatusCode), err), resp.StatusCode == http.StatusServiceUnavailable
	}
	if len(data) > 4096 {
		return nil, errors.Join(errors.New("response exceeds 4096 bytes"), err), false
	}
	if err != nil {
		return nil, err, readErr != nil
	}
	return data, nil, false
}

// Observe causes before net/http's client timeout wrappers replace them.
type observedTransport struct {
	next         http.RoundTripper
	transportErr error
	bodyErr      error
	stopErr      error
	ctx          context.Context
	status       int
}

func (t *observedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.ctx = req.Context()
	resp, err := t.next.RoundTrip(req)
	t.transportErr = err
	if err != nil {
		t.observeStop()
	}
	if err == nil && resp != nil {
		t.status = resp.StatusCode
		if resp.Body != nil {
			resp.Body = &observedBody{ReadCloser: resp.Body, transport: t}
		}
	}
	return resp, err
}

func (t *observedTransport) observeStop() {
	if err := t.ctx.Err(); err != nil {
		t.stopErr = errors.Join(err, context.Cause(t.ctx))
	} else if deadline, ok := t.ctx.Deadline(); ok && !time.Now().Before(deadline) {
		// Client's legacy cancellation timer can fire before the context timer.
		t.stopErr = context.DeadlineExceeded
	}
}

func (t *observedTransport) CancelRequest(req *http.Request) {
	if canceler, ok := t.next.(interface{ CancelRequest(*http.Request) }); ok {
		canceler.CancelRequest(req)
	}
}

type observedBody struct {
	io.ReadCloser
	transport *observedTransport
}

func (b *observedBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err != nil && err != io.EOF {
		b.transport.bodyErr = err
		b.transport.observeStop()
	}
	return n, err
}
