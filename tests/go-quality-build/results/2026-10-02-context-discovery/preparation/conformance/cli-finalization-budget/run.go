package finalize

import "context"
import "errors"
import "time"

// Run applies jobs sequentially and finalizes their accepted count.
func Run(ctx context.Context, jobs []int, apply func(context.Context, int) error, finalize func(context.Context, int) error, finalizationBudget time.Duration) (accepted int, err error) {
	if err := ctx.Err(); err != nil {
		return 0, errors.Join(err, context.Cause(ctx))
	}
	for _, job := range jobs {
		if err = ctx.Err(); err != nil {
			err = errors.Join(err, context.Cause(ctx))
			break
		}
		if err = apply(ctx, job); err != nil {
			err = errors.Join(err, ctx.Err(), context.Cause(ctx))
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
