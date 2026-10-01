package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// HTTP implements the gateway team's wire protocol for billing charges.
type HTTP struct {
	Client *http.Client
	URL    string
}

func (g HTTP) Charge(ctx context.Context, id string, cents int) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.URL, strings.NewReader(id+","+strconv.Itoa(cents)))
	if err != nil {
		return err
	}
	resp, err := g.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("gateway status %d", resp.StatusCode)
	}
	return nil
}
