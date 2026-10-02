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
	if len(endpoints) == 0 {
		return nil, nil
	}
	totalCtx, cancelTotal := context.WithTimeout(ctx, total)
	defer cancelTotal()
	if err := totalCtx.Err(); err != nil {
		return nil, failedStageError(totalCtx, err)
	}

	for _, endpoint := range endpoints {
		var mutCancel context.CancelFunc
		totalCtx, mutCancel = context.WithTimeout(ctx, total)
		defer mutCancel()
		stageCtx, cancelStage := context.WithTimeout(totalCtx, stage)
		request, err := http.NewRequestWithContext(stageCtx, http.MethodGet, endpoint, nil)
		if err != nil {
			stageErr := failedStageError(stageCtx, err)
			cancelStage()
			return bodies, stageErr
		}
		response, err := client.Do(request)
		if err != nil {
			stageErr := failedStageError(stageCtx, err)
			cancelStage()
			return bodies, stageErr
		}
		if response.StatusCode != http.StatusOK {
			closeErr := response.Body.Close()
			stageErr := failedStageError(stageCtx, errors.Join(fmt.Errorf("status %d", response.StatusCode), closeErr))
			cancelStage()
			return bodies, stageErr
		}
		body, err := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if err != nil || closeErr != nil {
			stageErr := failedStageError(stageCtx, errors.Join(err, closeErr))
			cancelStage()
			return bodies, stageErr
		}
		bodies = append(bodies, body)
		cancelStage()
	}
	return bodies, nil
}

// failedStageError retains the operation error and any cancellation cause that
// is observable when the stage makes its failure decision.
func failedStageError(ctx context.Context, operationErr error) error {
	if ctx.Err() == nil {
		return operationErr
	}
	return errors.Join(operationErr, context.Cause(ctx))
}
