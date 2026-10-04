package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	maxAttempts     = 3
	bodyLimit       = 4096
	operationBudget = 250 * time.Millisecond
	fallbackDelay   = 10 * time.Millisecond
)

// Fetch gets a complete response of at most 4096 bytes using the supplied client.
// It accepts HTTP 200 and retries only 429 and 503, up to three application
// attempts sharing 250ms or the earlier caller deadline. The caller owns client.
func Fetch(ctx context.Context, client *http.Client, endpoint string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, operationBudget)
	defer cancel()
	var last error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return nil, stoppedError(ctx, last, nil)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, boundedError("request failed", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			// On a Do error, any redirect response body is already closed.
			return nil, stoppedError(ctx, last, err)
		}
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, bodyLimit+1))
		closeErr := resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if len(data) > bodyLimit {
				return nil, stoppedError(ctx, nil, boundedError("response exceeds 4096 bytes", readErr, closeErr))
			}
			if readErr != nil || closeErr != nil {
				return nil, stoppedError(ctx, nil, boundedError("response failed", readErr, closeErr))
			}
			return data, nil
		}
		if len(data) > bodyLimit {
			data = data[:bodyLimit]
		}
		message := fmt.Sprintf("HTTP %d", resp.StatusCode)
		if len(data) != 0 {
			message += ": " + string(data)
		}
		last = boundedError(message, readErr, closeErr)
		if readErr != nil || closeErr != nil {
			return nil, stoppedError(ctx, last, nil)
		}
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable || attempt == maxAttempts-1 {
			return nil, last
		}
		if ctx.Err() != nil {
			return nil, stoppedError(ctx, last, nil)
		}
		delay, representable := retryDelay(resp.Header.Get("Retry-After"), time.Now())
		deadline, _ := ctx.Deadline()
		if !representable || delay >= time.Until(deadline) {
			return nil, last
		}
		if delay != 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, stoppedError(ctx, last, nil)
			case <-timer.C:
			}
		}
	}
	return nil, last
}

func retryDelay(value string, now time.Time) (time.Duration, bool) {
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
		if err != nil || seconds > uint64((1<<63-1)/int64(time.Second)) {
			return 0, false
		}
		return time.Duration(seconds) * time.Second, true
	}
	if date, err := http.ParseTime(value); err == nil {
		if !date.After(now) {
			return 0, true
		}
		return date.Sub(now), true
	}
	return fallbackDelay, true
}

func stoppedError(ctx context.Context, last, failure error) error {
	message := "request failed"
	if last != nil {
		message = last.Error()
	}
	if err := ctx.Err(); err != nil {
		return boundedError(message, last, failure, err, context.Cause(ctx))
	}
	return boundedError(message, last, failure)
}

type fetchError struct {
	message string
	causes  []error
}

func (e *fetchError) Error() string   { return e.message }
func (e *fetchError) Unwrap() []error { return e.causes }

func boundedError(message string, causes ...error) error {
	if len(message) > bodyLimit {
		message = message[:bodyLimit]
	}
	e := &fetchError{message: message}
	for _, cause := range causes {
		if cause == nil {
			continue
		}
		e.causes = append(e.causes, cause)
		if len(e.message) < bodyLimit {
			text := ": " + cause.Error()
			if remaining := bodyLimit - len(e.message); len(text) > remaining {
				text = text[:remaining]
			}
			e.message += text
		}
	}
	return e
}
