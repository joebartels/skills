package process

import (
	"context"
	"errors"
)

// Process applies jobs in order and returns the number completed before failure.
func Process(ctx context.Context, jobs []int, apply func(context.Context, int) error) (accepted int, err error) {
	if err := cancellationError(ctx); err != nil {
		return 0, err
	}

	for _, job := range jobs {
		if err := cancellationError(ctx); err != nil {
			return accepted, err
		}
		if callbackErr := apply(ctx, job); callbackErr != nil {
			if cancelErr := cancellationError(ctx); cancelErr != nil {
				return accepted, errors.Join(callbackErr, cancelErr)
			}
			return accepted, callbackErr
		}
		accepted++
	}
	return accepted, nil
}

func cancellationError(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return errors.Join(err, context.Cause(ctx))
	}
	return nil
}
