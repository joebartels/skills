package process

import "context"
import "errors"

// Process applies jobs in order and returns the number completed before failure.
func Process(ctx context.Context, jobs []int, apply func(context.Context, int) error) (accepted int, err error) {
	if err := ctx.Err(); err != nil {
		return 0, errors.Join(err, context.Cause(ctx))
	}
	for _, job := range jobs {
		if err := ctx.Err(); err != nil {
			return accepted, errors.Join(err, context.Cause(ctx))
		}
		if err := apply(ctx, job); err != nil {
			return accepted, errors.Join(err, ctx.Err(), context.Cause(ctx))
		}
		accepted++
	}
	return accepted, nil
}
