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
	var result error
	for {
		cohort := make([]Job, 0, limit)
		for len(cohort) < limit {
			if err := ctx.Err(); err != nil {
				result = errors.Join(result, err)
				return result
			}
			select {
			case <-ctx.Done():
				result = errors.Join(result, ctx.Err())
				return result
			case job, ok := <-jobs:
				if !ok {
					if len(cohort) == 0 {
						return result
					}
					goto runCohort
				}
				cohort = append(cohort, job)
			}
		}

	runCohort:
		workCtx, cancel := context.WithCancel(ctx)
		type opened struct {
			lease Lease
			err   error
			job   Job
		}
		openedCh := make(chan opened, len(cohort))
		for _, job := range cohort {
			go func(job Job) {
				lease, err := open(workCtx, job)
				openedCh <- opened{lease: lease, err: err, job: job}
			}(job)
		}
		leases := make([]Lease, 0, len(cohort))
		runs := make(chan error, len(cohort))
		runCount := 0
		for range cohort {
			opened := <-openedCh
			if opened.err != nil {
				result = errors.Join(result, opened.err)
				cancel()
			}
			if opened.lease != nil {
				leases = append(leases, opened.lease)
				if opened.err == nil && workCtx.Err() == nil {
					runCount++
					go func(lease Lease, job Job) { runs <- lease.Run(workCtx, job) }(opened.lease, opened.job)
				}
			}
		}
		if result != nil || ctx.Err() != nil {
			cancel()
		}
		if ctx.Err() != nil {
			result = errors.Join(result, ctx.Err())
		}
		for i := 0; i < runCount; i++ {
			if err := <-runs; err != nil {
				result = errors.Join(result, err)
				cancel()
			}
		}
		cancel()
		for _, lease := range leases {
			if err := lease.Close(); err != nil {
				result = errors.Join(result, err)
			}
		}
		if ctx.Err() != nil {
			result = errors.Join(result, ctx.Err())
		}
		if result != nil {
			return result
		}
	}
}
