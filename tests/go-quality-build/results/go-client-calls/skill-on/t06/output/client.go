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

// Submit posts payload within 500ms and accepts only HTTP 200. The caller owns
// client. With deduplication, operationID must identify the same payload across
// caller retries; without it, a transport failure may follow a completed effect.
func Submit(ctx context.Context, client *http.Client, endpoint, operationID string, payload []byte, deduplicates bool) ([]byte, error) {
	if deduplicates && operationID == "" {
		return nil, errors.New("empty operation identity")
	}
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	payload = bytes.Clone(payload)

	// A redirect is a terminal status, even if its target would return 200.
	callClient := *client
	// Request scopes retain the timeout without Go 1.22's body-error masking.
	callClient.Timeout = 0
	callClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if client.CheckRedirect != nil {
			if err := client.CheckRedirect(req, via); err != nil {
				return err
			}
		}
		return http.ErrUseLastResponse
	}
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, errors.Join(lastErr, err)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		if deduplicates {
			req.Header.Set("Idempotency-Key", operationID)
		}
		data, retry, err := submitAttempt(&callClient, req, client.Timeout)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if err := ctx.Err(); err != nil {
			return nil, errors.Join(lastErr, err)
		}
		if !deduplicates || !retry || attempt == 2 {
			return nil, lastErr
		}
		timer := time.NewTimer(time.Duration(attempt+1) * 10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, errors.Join(lastErr, ctx.Err())
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func submitAttempt(client *http.Client, req *http.Request, timeout time.Duration) ([]byte, bool, error) {
	ctx := req.Context()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}
	transport := &statusTransport{base: client.Transport}
	if transport.base == nil {
		transport.base = http.DefaultTransport
	}
	attemptClient := *client
	attemptClient.Transport = transport
	resp, err := attemptClient.Do(req)
	if err != nil {
		// On redirect-policy failure, Do has already closed resp.Body.
		if resp != nil {
			return nil, false, errors.Join(fmt.Errorf("HTTP %d", resp.StatusCode), err, ctx.Err())
		}
		// An invalid redirect Location hides the response before CheckRedirect.
		if transport.status != 0 {
			return nil, false, errors.Join(fmt.Errorf("HTTP %d", transport.status), err, ctx.Err())
		}
		return nil, true, errors.Join(err, ctx.Err())
	}
	data, retry, err := readResponse(resp)
	if err != nil {
		return nil, retry, errors.Join(err, ctx.Err())
	}
	return data, false, nil
}

type statusTransport struct {
	base   http.RoundTripper
	status int
}

func (t *statusTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err == nil && resp != nil {
		t.status = resp.StatusCode
	}
	return resp, err
}

func readResponse(resp *http.Response) ([]byte, bool, error) {
	defer resp.Body.Close()
	limit := int64(4096)
	if resp.StatusCode == http.StatusOK {
		limit++ // One extra byte distinguishes a complete result from truncation.
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode == http.StatusServiceUnavailable,
			errors.Join(fmt.Errorf("HTTP %d: %s", resp.StatusCode, data), err)
	}
	if len(data) > 4096 {
		return nil, false, errors.Join(errors.New("response exceeds 4096 bytes"), err)
	}
	return data, err != nil, err
}
