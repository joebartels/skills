package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type Client struct{ http *http.Client }

func New(client *http.Client, enabled bool) *Client { return &Client{http: client} }
func (c *Client) Fetch(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
