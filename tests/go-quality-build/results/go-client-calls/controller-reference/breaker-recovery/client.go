package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/sony/gobreaker/v2"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Client struct {
	http     *http.Client
	enabled  bool
	mu       sync.Mutex
	breakers map[string]*gobreaker.CircuitBreaker[[]byte]
}

func New(client *http.Client, enabled bool) *Client {
	return &Client{http: client, enabled: enabled, breakers: map[string]*gobreaker.CircuitBreaker[[]byte]{}}
}

type healthFailure struct{ error }

func (c *Client) Fetch(ctx context.Context, endpoint string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	work := func() ([]byte, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		b, err := consume(resp)
		if resp.StatusCode == 503 && ctx.Err() == nil {
			return b, healthFailure{err}
		}
		return b, err
	}
	if !c.enabled {
		return work()
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	key := u.Scheme + "://" + u.Host
	c.mu.Lock()
	cb := c.breakers[key]
	if cb == nil {
		cb = gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{MaxRequests: 1, Timeout: 100 * time.Millisecond, ReadyToTrip: func(c gobreaker.Counts) bool { return c.ConsecutiveFailures >= 3 }, IsExcluded: func(err error) bool { var h healthFailure; return err != nil && !errors.As(err, &h) }})
		c.breakers[key] = cb
	}
	c.mu.Unlock()
	return cb.Execute(work)
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
