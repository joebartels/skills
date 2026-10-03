package stages

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

func FetchAll(ctx context.Context, client *http.Client, endpoints []string, total, stage time.Duration) (bodies [][]byte, err error) {
	op, cancel := context.WithTimeout(ctx, total)
	defer cancel()
	for _, endpoint := range endpoints {
		if err := op.Err(); err != nil {
			return bodies, errors.Join(err, context.Cause(op))
		}
		body, err := fetchStage(op, client, endpoint, stage)
		if err != nil {
			return bodies, err
		}
		bodies = append(bodies, body)
	}
	return bodies, nil
}
func fetchStage(parent context.Context, client *http.Client, endpoint string, stage time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(parent, stage)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, errors.Join(err, ctx.Err(), context.Cause(ctx))
	}
	if response.StatusCode != http.StatusOK {
		err := errors.Join(fmt.Errorf("status %d", response.StatusCode), response.Body.Close())
		return nil, errors.Join(err, ctx.Err(), context.Cause(ctx))
	}
	body, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return nil, errors.Join(err, ctx.Err(), context.Cause(ctx))
	}
	return body, nil
}
