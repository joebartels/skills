package finalize

import (
	"context"
	"errors"
	"time"
)

// Run applies jobs sequentially and finalizes their accepted count.
func Run(ctx context.Context, jobs []int, apply func(context.Context, int) error, finalize func(context.Context, int) error, finalizationBudget time.Duration) (accepted int, err error) {
	for _, job := range jobs {
		if ctxErr := ctx.Err(); ctxErr != nil {
			err = cancellationError(ctx)
			break
		}
		if applyErr := apply(ctx, job); applyErr != nil {
			err = applyErr
			if ctx.Err() != nil {
				err = errors.Join(err, cancellationError(ctx))
			}
			break
		}
		accepted++
	}
	if accepted != 0 {
		finalCtx, cancel := context.WithTimeout(context.Background(), finalizationBudget)
		finalErr := finalize(finalCtx, accepted)
		cancel()
		err = errors.Join(err, finalErr)
	}
	return accepted, err
}

func cancellationError(ctx context.Context) error {
	return errors.Join(ctx.Err(), context.Cause(ctx))
}
