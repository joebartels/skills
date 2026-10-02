package finalize

import "context"
import "time"

// Run applies jobs sequentially and finalizes their accepted count.
func Run(ctx context.Context, jobs []int, apply func(context.Context, int) error, finalize func(context.Context, int) error, finalizationBudget time.Duration) (accepted int, err error) {
	for _, job := range jobs {
		if err = apply(ctx, job); err != nil {
			break
		}
		accepted++
	}
	if accepted != 0 {
		finalErr := finalize(ctx, accepted)
		if err == nil {
			err = finalErr
		}
	}
	return accepted, err
}
