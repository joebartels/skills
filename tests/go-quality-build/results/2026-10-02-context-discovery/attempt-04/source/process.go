package process

import (
	"context"
	"errors"
)

// Process applies jobs in order and returns the number completed before failure.
func Process(ctx context.Context, jobs []int, apply func(context.Context, int) error) (accepted int, err error) {
	for _, job := range jobs {
		if ctx.Err() != nil {
			return accepted, cancellationError(ctx)
		}
		if callbackErr := apply(ctx, job); callbackErr != nil {
			if ctx.Err() != nil {
				return accepted, errors.Join(callbackErr, cancellationError(ctx))
			}
			return accepted, callbackErr
		}
		accepted++
	}
	return accepted, nil
}

func cancellationError(ctx context.Context) error {
	return errors.Join(ctx.Err(), context.Cause(ctx))
}
