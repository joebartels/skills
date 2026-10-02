package process

import "context"

// Process applies jobs in order and returns the number completed before failure.
func Process(ctx context.Context, jobs []int, apply func(context.Context, int) error) (accepted int, err error) {
	for _, job := range jobs {
		if err := apply(ctx, job); err != nil {
			return accepted, err
		}
		accepted++
	}
	return accepted, nil
}
