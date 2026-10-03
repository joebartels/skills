package stages

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FetchAll obtains complete response bodies in endpoint order.
func FetchAll(ctx context.Context, client *http.Client, endpoints []string, total, stage time.Duration) (bodies [][]byte, err error) {
	for _, endpoint := range endpoints {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return bodies, err
		}
		response, err := client.Do(request)
		if err != nil {
			return bodies, err
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			return bodies, fmt.Errorf("status %d", response.StatusCode)
		}
		body, err := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if err != nil {
			return bodies, err
		}
		if closeErr != nil {
			return bodies, closeErr
		}
		bodies = append(bodies, body)
	}
	return bodies, nil
}
