package finalize

import (
	"context"
	"errors"
	"time"
)

// Run applies jobs sequentially and finalizes their accepted count.
func Run(ctx context.Context, jobs []int, apply func(context.Context, int) error, finalize func(context.Context, int) error, finalizationBudget time.Duration) (accepted int, err error) {
	for _, job := range jobs {
		if err = ctx.Err(); err != nil {
			err = observedCancellation(ctx, err)
			break
		}
		if applyErr := apply(ctx, job); applyErr != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				err = observedCancellation(ctx, ctxErr)
				if !errors.Is(applyErr, ctxErr) {
					err = errors.Join(applyErr, err)
				}
			} else {
				err = applyErr
			}
			break
		}
		accepted++
	}
	if accepted != 0 {
		finalCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), finalizationBudget)
		finalErr := finalize(finalCtx, accepted)
		cancel()
		err = errors.Join(err, finalErr)
	}
	return accepted, err
}

func observedCancellation(ctx context.Context, classification error) error {
	return errors.Join(classification, context.Cause(ctx))
}
