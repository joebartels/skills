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

// Fetch returns a complete response of at most 4096 bytes from an HTTP 200.
// It borrows client and retries only 429 and 503, up to three attempts within
// 250ms or the earlier context deadline, honoring Retry-After when it fits.
func Fetch(ctx context.Context, client *http.Client, endpoint string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		if ctx.Err() != nil {
			return nil, stoppedError(ctx, last)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, newFetchError("", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			message := ""
			if resp != nil {
				message = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
			// On a redirect-policy error, Client.Do already closed the body.
			return nil, stoppedError(ctx, newFetchError(message, last, err))
		}
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, bodyLimit+1))
		closeErr := resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if len(data) > bodyLimit {
				return nil, stoppedError(ctx, newFetchError("response exceeds 4096 bytes", readErr, closeErr))
			}
			if err := errors.Join(readErr, closeErr); err != nil {
				return nil, stoppedError(ctx, newFetchError("", err))
			}
			return data, nil
		}
		message := fmt.Sprintf("HTTP %d", resp.StatusCode)
		if len(data) > 0 {
			message += ": " + string(data[:min(len(data), bodyLimit)])
		}
		last = newFetchError(message, readErr, closeErr)
		if readErr != nil || closeErr != nil || attempt == 2 ||
			(resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable) {
			return nil, stoppedError(ctx, last)
		}
		if ctx.Err() != nil {
			return nil, stoppedError(ctx, last)
		}
		delay, representable := retryDelay(resp.Header.Get("Retry-After"), time.Now())
		deadline, _ := ctx.Deadline()
		if !representable || delay >= time.Until(deadline) {
			return nil, stoppedError(ctx, last)
		}
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, stoppedError(ctx, last)
			case <-timer.C:
			}
		}
	}
	return nil, last
}

func retryDelay(hint string, now time.Time) (time.Duration, bool) {
	hint = strings.TrimSpace(hint)
	digits := hint != ""
	for _, c := range hint {
		if c < '0' || c > '9' {
			digits = false
			break
		}
	}
	if digits {
		seconds, err := strconv.ParseUint(hint, 10, 64)
		if err != nil || seconds > uint64(time.Duration(1<<63-1)/time.Second) {
			return 0, false
		}
		return time.Duration(seconds) * time.Second, true
	}
	if date, err := http.ParseTime(hint); err == nil {
		return max(0, date.Sub(now)), true
	}
	return 10 * time.Millisecond, true
}

func stoppedError(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return newFetchError("", err, ctxErr, context.Cause(ctx))
	}
	return err
}

type fetchError struct {
	message string
	cause   error
}

func (e *fetchError) Error() string { return e.message }
func (e *fetchError) Unwrap() error { return e.cause }

func newFetchError(message string, causes ...error) error {
	var text strings.Builder
	appendText := func(s string) {
		text.WriteString(s[:min(len(s), bodyLimit-text.Len())])
	}
	appendText(message)
	for _, cause := range causes {
		if cause == nil || text.Len() == bodyLimit {
			continue
		}
		if text.Len() > 0 {
			appendText(": ")
		}
		appendText(cause.Error())
	}
	return &fetchError{message: text.String(), cause: errors.Join(causes...)}
}
