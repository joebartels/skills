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

func Submit(ctx context.Context, client *http.Client, endpoint, operationID string, payload []byte, deduplicates bool) ([]byte, error) {
	if deduplicates && operationID == "" {
		return nil, errors.New("empty identity")
	}
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	attempts := 1
	if deduplicates {
		attempts = 3
	}
	var last error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return nil, errors.Join(last, err)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		if deduplicates {
			req.Header.Set("Idempotency-Key", operationID)
		}
		resp, err := client.Do(req)
		eligible := true
		if err == nil {
			var b []byte
			b, err = consume(resp)
			if err == nil {
				return b, nil
			}
			eligible = resp.StatusCode == 503
		}
		last = err
		if !eligible || i == attempts-1 {
			return nil, last
		}
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, errors.Join(last, ctx.Err())
		case <-timer.C:
		}
	}
	return nil, last
}

func consume(resp *http.Response) ([]byte, error) {
	b, readErr := io.ReadAll(io.LimitReader(resp.Body, 4097))
	closeErr := resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.Join(fmt.Errorf("HTTP %d: %s", resp.StatusCode, b[:min(len(b), 4096)]), readErr, closeErr)
	}
	if len(b) > 4096 {
		return nil, errors.Join(errors.New("response exceeds limit"), readErr, closeErr)
	}
	return b, errors.Join(readErr, closeErr)
}
