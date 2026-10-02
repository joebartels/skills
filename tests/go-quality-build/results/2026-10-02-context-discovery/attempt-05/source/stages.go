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
	totalCtx, cancelTotal := context.WithTimeout(ctx, total)
	defer cancelTotal()

	for _, endpoint := range endpoints {
		stageCtx, cancelStage := context.WithTimeout(totalCtx, stage)
		if stageCtx.Err() != nil {
			stageErr := stageError(stageCtx.Err(), stageCtx)
			cancelStage()
			return bodies, stageErr
		}
		request, requestErr := http.NewRequestWithContext(stageCtx, http.MethodGet, endpoint, nil)
		if requestErr != nil {
			cancelStage()
			return bodies, requestErr
		}
		response, requestErr := client.Do(request)
		if requestErr != nil {
			requestErr = stageError(requestErr, stageCtx)
			cancelStage()
			return bodies, requestErr
		}

		if response.StatusCode != http.StatusOK {
			closeErr := response.Body.Close()
			stageErr := stageError(errors.Join(fmt.Errorf("status %d", response.StatusCode), closeErr), stageCtx)
			cancelStage()
			return bodies, stageErr
		}
		body, readErr := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if readErr != nil || closeErr != nil {
			stageErr := stageError(errors.Join(readErr, closeErr), stageCtx)
			cancelStage()
			return bodies, stageErr
		}
		bodies = append(bodies, body)
		cancelStage()
	}
	return bodies, nil
}

// stageError preserves operation failures and the cancellation state observed
// at the stage's failure decision, including a custom context cause.
func stageError(operationErr error, ctx context.Context) error {
	if operationErr == nil {
		return nil
	}
	return errors.Join(operationErr, ctx.Err(), context.Cause(ctx))
}
