package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sony/gobreaker/v2"
)

// Client fetches bounded responses using a borrowed HTTP client.
type Client struct {
	http    *http.Client
	enabled bool

	mu       sync.Mutex
	breakers map[dependency]*gobreaker.CircuitBreaker[[]byte]
}

type dependency struct {
	scheme    string
	authority string
}

type statusError struct {
	status   int
	eligible bool
}

func (e *statusError) Error() string { return fmt.Sprintf("HTTP %d", e.status) }

// New borrows client and optionally enables a breaker for each dependency.
// The caller retains ownership of the nonnil client and its configuration.
func New(client *http.Client, enabled bool) *Client {
	return &Client{
		http:     client,
		enabled:  enabled,
		breakers: make(map[dependency]*gobreaker.CircuitBreaker[[]byte]),
	}
}

// Fetch makes one request under a 250ms budget and returns at most 4096 bytes.
// Only HTTP 200 succeeds; redirects are returned as HTTP status errors.
func (c *Client) Fetch(ctx context.Context, endpoint string) ([]byte, error) {
	requestCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	if err := requestCtx.Err(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if !c.enabled {
		return c.fetch(req)
	}
	key := dependency{scheme: strings.ToLower(req.URL.Scheme), authority: strings.ToLower(req.URL.Host)}
	return c.breaker(key).Execute(func() ([]byte, error) {
		data, err := c.fetch(req)
		if status, ok := err.(*statusError); ok {
			status.eligible = status.status == http.StatusServiceUnavailable && ctx.Err() == nil
		}
		return data, err
	})
}

func (c *Client) breaker(key dependency) *gobreaker.CircuitBreaker[[]byte] {
	c.mu.Lock()
	defer c.mu.Unlock()
	if breaker := c.breakers[key]; breaker != nil {
		return breaker
	}
	breaker := gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
		Name:        key.scheme + "://" + key.authority,
		MaxRequests: 1,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts gobreaker.Counts) bool { return counts.ConsecutiveFailures >= 3 },
		IsExcluded: func(err error) bool {
			if err == nil {
				return false
			}
			status, ok := err.(*statusError)
			return !ok || !status.eligible
		},
	})
	c.breakers[key] = breaker
	return breaker
}

func (c *Client) fetch(req *http.Request) ([]byte, error) {
	// A copy keeps the caller's policy intact while preventing a second request.
	client := *c.http
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &statusError{status: resp.StatusCode}
	}
	return io.ReadAll(io.LimitReader(resp.Body, 4096))
}
