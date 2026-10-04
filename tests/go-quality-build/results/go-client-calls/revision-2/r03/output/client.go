package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sony/gobreaker/v2"
)

const (
	fetchBudget   = 250 * time.Millisecond
	responseLimit = 4096
)

var (
	errExcluded    = errors.New("excluded health outcome")
	errUnavailable = errors.New("unhealthy dependency")
)

type dependency struct{ scheme, authority string }

// Client fetches bounded responses using a borrowed HTTP client.
type Client struct {
	http     *http.Client
	mu       sync.Mutex
	breakers map[dependency]*gobreaker.TwoStepCircuitBreaker[[]byte]
}

// New returns a client with optional breakers scoped to each dependency.
// A nil HTTP client uses a new default client.
func New(client *http.Client, enabled bool) *Client {
	if client == nil {
		client = &http.Client{}
	}
	c := &Client{http: client}
	if enabled {
		c.breakers = make(map[dependency]*gobreaker.TwoStepCircuitBreaker[[]byte])
	}
	return c
}

// Fetch makes one application request with a total budget of 250ms.
// Only HTTP 200 succeeds; complete responses may contain at most 4096 bytes.
func (c *Client) Fetch(ctx context.Context, endpoint string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, fetchBudget)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, boundError(err)
	}
	if c.breakers == nil {
		data, _, err := c.fetch(req)
		return data, err
	}
	done, err := c.breaker(req.URL).Allow()
	if err != nil {
		return nil, err
	}
	data, status, err := c.fetch(req)
	// Cancellation is neutral even if the response completed successfully.
	if ctx.Err() != nil {
		done(errExcluded)
	} else if status == http.StatusServiceUnavailable {
		done(errUnavailable)
	} else if err == nil {
		done(nil)
	} else {
		done(errExcluded)
	}
	return data, err
}

func (c *Client) breaker(endpoint *url.URL) *gobreaker.TwoStepCircuitBreaker[[]byte] {
	key := dependency{strings.ToLower(endpoint.Scheme), strings.ToLower(endpoint.Host)}
	c.mu.Lock()
	defer c.mu.Unlock()
	if breaker := c.breakers[key]; breaker != nil {
		return breaker
	}
	breaker := gobreaker.NewTwoStepCircuitBreaker[[]byte](gobreaker.Settings{
		MaxRequests: 1,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts gobreaker.Counts) bool { return counts.ConsecutiveFailures >= 3 },
		IsExcluded:  func(err error) bool { return errors.Is(err, errExcluded) },
	})
	c.breakers[key] = breaker
	return breaker
}

func (c *Client) fetch(req *http.Request) ([]byte, int, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, boundError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, responseLimit+1))
	if len(data) > responseLimit {
		return nil, resp.StatusCode, boundError(errors.Join(errors.New("response exceeds 4096 byte limit"), err))
	}
	return data, resp.StatusCode, boundError(err)
}

type diagnosticError struct {
	message string
	cause   error
}

func (e *diagnosticError) Error() string { return e.message }
func (e *diagnosticError) Unwrap() error { return e.cause }

func boundError(err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	if len(message) <= responseLimit {
		return err
	}
	return &diagnosticError{message: message[:responseLimit], cause: err}
}
