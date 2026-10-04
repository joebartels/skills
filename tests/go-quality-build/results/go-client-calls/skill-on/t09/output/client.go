package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	healthExcluded    = errors.New("excluded health outcome")
	healthUnavailable = errors.New("unavailable health outcome")
)

type dependency struct {
	scheme    string
	authority string
}

type Client struct {
	http     *http.Client
	mu       sync.Mutex
	breakers map[dependency]*gobreaker.TwoStepCircuitBreaker[struct{}]
}

// New borrows client and optionally enables a breaker for each endpoint dependency.
func New(client *http.Client, enabled bool) *Client {
	c := &Client{http: client}
	if enabled {
		c.breakers = make(map[dependency]*gobreaker.TwoStepCircuitBreaker[struct{}])
	}
	return c
}

// Fetch makes one application request with a 250ms budget and a bounded result.
// The supplied HTTP client's redirect policy applies to that request.
func (c *Client) Fetch(ctx context.Context, endpoint string) ([]byte, error) {
	operationCtx, cancel := context.WithTimeout(ctx, fetchBudget)
	defer cancel()
	if err := operationCtx.Err(); err != nil {
		return nil, boundedError(err)
	}
	req, err := http.NewRequestWithContext(operationCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, boundedError(err)
	}
	var done func(error)
	if c.breakers != nil {
		key := dependency{scheme: strings.ToLower(req.URL.Scheme), authority: strings.ToLower(req.URL.Host)}
		done, err = c.breaker(key).Allow()
		if err != nil {
			return nil, boundedError(err)
		}
	}
	data, status, err := c.fetch(req)
	if done != nil {
		// Caller stopping is neutral even when the request returned a completed result.
		health := healthExcluded
		if operationCtx.Err() == nil {
			if status == http.StatusServiceUnavailable {
				health = healthUnavailable
			} else if status == http.StatusOK && err == nil {
				health = nil
			}
		}
		done(health)
	}
	return data, boundedError(err)
}

func (c *Client) breaker(key dependency) *gobreaker.TwoStepCircuitBreaker[struct{}] {
	c.mu.Lock()
	defer c.mu.Unlock()
	if b := c.breakers[key]; b != nil {
		return b
	}
	b := gobreaker.NewTwoStepCircuitBreaker[struct{}](gobreaker.Settings{
		MaxRequests: 1,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts gobreaker.Counts) bool { return counts.ConsecutiveFailures >= 3 },
		IsExcluded:  func(err error) bool { return err == healthExcluded },
	})
	c.breakers[key] = b
	return b
}

func (c *Client) fetch(req *http.Request) ([]byte, int, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		// Do's error response, if any, already has a closed body.
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, responseLimit+1))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if len(data) > responseLimit {
		return nil, resp.StatusCode, errors.New("response exceeds 4096 bytes")
	}
	return data, resp.StatusCode, nil
}

type diagnosticError struct {
	message string
	cause   error
}

func (e *diagnosticError) Error() string { return e.message }
func (e *diagnosticError) Unwrap() error { return e.cause }

func boundedError(err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	if len(message) <= responseLimit {
		return err
	}
	return &diagnosticError{message: message[:responseLimit], cause: err}
}
