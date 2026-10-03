package workers

import (
	"context"
	"errors"
	"sync"
)

// Job identifies a requested operation.
type Job int

// Lease supplies the existing cooperative operation and release boundary.
type Lease interface {
	Run(context.Context, Job) error
	Close() error
}

type runResult struct {
	lease Lease
	err   error
}

// Serve runs supplied jobs using leases acquired by the host.
func Serve(ctx context.Context, jobs <-chan Job, limit int, open func(context.Context, Job) (Lease, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var errs []error
	for ctx.Err() == nil {
		cohortCtx, stopCohort := context.WithCancel(ctx)
		cohort := make([]Lease, 0, limit)
		runs := make(chan runResult, limit)
		var wg sync.WaitGroup
		var admissionErr error

		// Admit immediately while capacity is available. At a full cohort,
		// collect completion before admitting more work.
		for len(cohort) < limit && admissionErr == nil {
			select {
			case <-ctx.Done():
				admissionErr = ctx.Err()
			case job, ok := <-jobs:
				if !ok {
					admissionErr = errJobsClosed
					break
				}
				lease, err := open(cohortCtx, job)
				if err != nil {
					admissionErr = err
					break
				}
				if lease == nil {
					admissionErr = errors.New("open returned nil lease without error")
					break
				}
				cohort = append(cohort, lease)
				wg.Add(1)
				go func(l Lease, j Job) {
					defer wg.Done()
					runs <- runResult{lease: l, err: l.Run(cohortCtx, j)}
				}(lease, job)
			}
		}

		inputClosed := admissionErr == errJobsClosed
		if inputClosed {
			admissionErr = nil
		}
		if admissionErr != nil {
			errs = append(errs, admissionErr)
		}
		if admissionErr != nil || ctx.Err() != nil {
			stopCohort()
		}

		completed := 0
		failed := admissionErr != nil || ctx.Err() != nil
		for completed < len(cohort) {
			select {
			case <-ctx.Done():
				if !failed {
					errs = append(errs, ctx.Err())
				}
				failed = true
			case result := <-runs:
				completed++
				if result.err != nil {
					errs = append(errs, result.err)
					if !failed {
						failed = true
						stopCohort()
					}
				}
			}
		}
		wg.Wait()
		stopCohort()
		for _, lease := range cohort {
			if err := lease.Close(); err != nil {
				errs = append(errs, err)
			}
		}

		if failed || ctx.Err() != nil {
			if ctx.Err() != nil {
				errs = append(errs, ctx.Err())
			}
			break
		}
		if len(cohort) == 0 {
			break
		}
		if inputClosed {
			break
		}
	}
	return errors.Join(errs...)
}

var errJobsClosed = errors.New("jobs closed")
