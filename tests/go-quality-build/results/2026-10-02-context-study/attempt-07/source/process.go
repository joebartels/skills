package process

import (
	"context"
	"errors"
)

// Process applies jobs in order and returns the number completed before failure.
func Process(ctx context.Context, jobs []int, apply func(context.Context, int) error) (accepted int, err error) {
	if ctx.Err() != nil {
		return 0, cancellationError(ctx)
	}
	for _, job := range jobs {
		if ctx.Err() != nil {
			return accepted, cancellationError(ctx)
		}
		if err := apply(ctx, job); err != nil {
			if ctx.Err() != nil {
				return accepted, errors.Join(err, cancellationError(ctx))
			}
			return accepted, err
		}
		accepted++
	}
	return accepted, nil
}

func cancellationError(ctx context.Context) error {
	return errors.Join(ctx.Err(), context.Cause(ctx))
}
