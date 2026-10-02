package stages

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FetchAll obtains complete response bodies in endpoint order.
func FetchAll(ctx context.Context, client *http.Client, endpoints []string, total, stage time.Duration) (bodies [][]byte, err error) {
	if ctx.Err() != nil {
		return nil, failedWithContext(ctx, nil)
	}
	totalCtx, cancelTotal := context.WithTimeout(ctx, total)
	defer cancelTotal()

	for _, endpoint := range endpoints {
		if err := totalCtx.Err(); err != nil {
			return bodies, failedWithContext(totalCtx, err)
		}
		body, err := fetchStage(totalCtx, client, endpoint, stage)
		if err != nil {
			return bodies, err
		}
		bodies = append(bodies, body)
	}
	return bodies, nil
}

func fetchStage(ctx context.Context, client *http.Client, endpoint string, budget time.Duration) ([]byte, error) {
	stageCtx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()
	request, err := http.NewRequestWithContext(stageCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, failedWithContext(stageCtx, err)
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, failedWithContext(stageCtx, err)
	}
	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("status %d", response.StatusCode)
	}
	var body []byte
	if err == nil {
		body, err = io.ReadAll(response.Body)
	}
	closeErr := response.Body.Close()
	if err != nil || closeErr != nil {
		return nil, failedWithContext(stageCtx, errors.Join(err, closeErr))
	}
	return body, nil
}

func failedWithContext(ctx context.Context, failure error) error {
	if ctx.Err() == nil {
		return failure
	}
	return errors.Join(failure, ctx.Err(), context.Cause(ctx))
}
