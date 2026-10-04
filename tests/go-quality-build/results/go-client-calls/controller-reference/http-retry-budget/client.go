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

func Fetch(ctx context.Context, client *http.Client, endpoint string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	var last error
	for i := 0; i < 3; i++ {
		if err := ctx.Err(); err != nil {
			return nil, errors.Join(last, err)
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		b, err := consume(resp)
		if err == nil {
			return b, nil
		}
		last = err
		if (resp.StatusCode != 429 && resp.StatusCode != 503) || i == 2 {
			return nil, last
		}
		delay, unfit := hint(resp.Header.Get("Retry-After"))
		deadline, _ := ctx.Deadline()
		if unfit || delay >= time.Until(deadline) {
			return nil, last
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, errors.Join(last, ctx.Err())
		case <-timer.C:
		}
	}
	return nil, last
}
func hint(s string) (time.Duration, bool) {
	s = strings.TrimSpace(s)
	digits := s != ""
	for _, r := range s {
		if r < '0' || r > '9' {
			digits = false
		}
	}
	if digits {
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil || n > uint64((1<<63-1)/int64(time.Second)) {
			return 0, true
		}
		return time.Duration(n) * time.Second, false
	}
	if d, err := http.ParseTime(s); err == nil {
		return max(time.Until(d), 0), false
	}
	return 10 * time.Millisecond, false
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
