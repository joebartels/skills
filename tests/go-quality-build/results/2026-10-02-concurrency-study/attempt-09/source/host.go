package workers

import (
	"context"
	"errors"
)

// Job identifies a requested operation.
type Job int

// Lease supplies the existing cooperative operation and release boundary.
type Lease interface {
	Run(context.Context, Job) error
	Close() error
}

// Serve runs supplied jobs using leases acquired by the host.
func Serve(ctx context.Context, jobs <-chan Job, limit int, open func(context.Context, Job) (Lease, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var failures []error
	inputClosed := false
	for !inputClosed {
		cohortCtx, stop := context.WithCancel(ctx)
		leases := make([]Lease, 0, limit)
		runResults := make(chan error, limit)
		started := 0

	admit:
		for len(leases) < limit {
			// Check before selecting too: a ready job must not win a select
			// after a stop has already been observed.
			if cohortCtx.Err() != nil {
				break admit
			}
			select {
			case <-cohortCtx.Done():
				break admit
			case job, ok := <-jobs:
				if !ok {
					inputClosed = true
					break admit
				}
				if cohortCtx.Err() != nil {
					break admit
				}
				lease, err := open(cohortCtx, job)
				if err != nil {
					failures = append(failures, err)
					stop()
					break admit
				}
				leases = append(leases, lease)
				if cohortCtx.Err() != nil {
					break admit
				}
				started++
				go func() {
					err := lease.Run(cohortCtx, job)
					if err != nil {
						stop()
					}
					runResults <- err
				}()
			}
		}

		// Every Run has returned before any lease in this cohort is closed.
		for i := 0; i < started; i++ {
			if err := <-runResults; err != nil {
				failures = append(failures, err)
			}
		}
		for _, lease := range leases {
			if err := lease.Close(); err != nil {
				failures = append(failures, err)
			}
		}
		stop()
		if len(failures) != 0 || ctx.Err() != nil {
			return errors.Join(append(failures, ctx.Err())...)
		}
	}
	return nil
}
