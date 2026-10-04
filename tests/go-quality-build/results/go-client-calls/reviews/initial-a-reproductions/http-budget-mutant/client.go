package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const bodyLimit = 4096

// Fetch reads up to 4096 bytes using client and a total 250ms budget.
// It retries HTTP 429 and 503 at most twice, honoring Retry-After.
func Fetch(ctx context.Context, client *http.Client, endpoint string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	var rejection error
	for attempt := 0; attempt < 3; attempt++ {
		if ctx.Err() != nil {
			return nil, fetchError(ctx, rejection, nil)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, fetchError(ctx, rejection, err)
		}
		resp, err := client.Do(req)
		if err != nil {
			// On a redirect error, Do has already closed its returned body.
			return nil, fetchError(ctx, rejection, err)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, bodyLimit+1))
		bodyErr := errors.Join(readErr, resp.Body.Close())
		if resp.StatusCode == http.StatusOK {
			if len(body) > bodyLimit {
				bodyErr = errors.Join(bodyErr, fmt.Errorf("HTTP 200 body exceeds %d bytes", bodyLimit))
			}
			if bodyErr != nil {
				return nil, fetchError(ctx, rejection, bodyErr)
			}
			return body, nil
		}
		if len(body) > bodyLimit {
			body = body[:bodyLimit]
		}
		rejection = fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
		if bodyErr != nil {
			return nil, fetchError(ctx, rejection, bodyErr)
		}
		if attempt == 2 || (resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable) {
			return nil, fetchError(ctx, rejection, nil)
		}
		delay := retryDelay(resp.Header.Get("Retry-After"), time.Now())
		deadline, _ := ctx.Deadline()
		if delay >= time.Until(deadline) {
			return nil, fetchError(ctx, rejection, nil)
		}
		if err := waitRetry(ctx, delay); err != nil {
			return nil, fetchError(ctx, rejection, err)
		}
	}
	return nil, rejection
}

func retryDelay(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	digits := value != ""
	for _, c := range value {
		if c < '0' || c > '9' {
			digits = false
			break
		}
	}
	if digits {
		seconds, err := strconv.ParseUint(value, 10, 64)
		const maxDuration = time.Duration(1<<63 - 1)
		if err != nil || seconds > uint64(maxDuration/time.Second) {
			return maxDuration
		}
		return time.Duration(seconds) * time.Second
	}
	if date, err := http.ParseTime(value); err == nil {
		if date.After(now) {
			return date.Sub(now)
		}
		return 0
	}
	return 10 * time.Millisecond
}

func waitRetry(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func fetchError(ctx context.Context, rejection, failure error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return errors.Join(rejection, failure, ctxErr, context.Cause(ctx))
	}
	return errors.Join(rejection, failure)
}
