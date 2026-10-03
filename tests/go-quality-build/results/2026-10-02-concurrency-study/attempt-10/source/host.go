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

// Serve runs supplied jobs using leases acquired by the host. It joins every
// Run in a cohort before closing any of that cohort's leases, and returns the
// independent Open, Run, Close, and caller cancellation errors it observes.
func Serve(ctx context.Context, jobs <-chan Job, limit int, open func(context.Context, Job) (Lease, error)) error {
	runCtx, stop := context.WithCancel(ctx)
	defer stop()

	var failures []error
	inputClosed := false
	for {
		// Acquired leases occupy capacity until the entire cohort has finished
		// running and every lease has been closed.
		leases := make([]Lease, 0, limit)
		finished := make(chan error, limit)
		running := 0
		stopping := false

		for len(leases) < limit && !inputClosed && !stopping {
			select {
			case <-runCtx.Done():
				stopping = true
			case err := <-finished:
				running--
				if err != nil {
					failures = append(failures, err)
					stopping = true
				}
			case job, ok := <-jobs:
				if !ok {
					inputClosed = true
					break
				}
				if runCtx.Err() != nil {
					stopping = true
					break
				}
				lease, err := open(runCtx, job)
				if err != nil {
					failures = append(failures, err)
					stopping = true
					stop()
					break
				}
				leases = append(leases, lease)
				if runCtx.Err() != nil {
					stopping = true
					break
				}
				running++
				go func() {
					err := lease.Run(runCtx, job)
					if err != nil {
						stop()
					}
					finished <- err
				}()
			}
		}

		// A failure in one Run requests a stop immediately, including while
		// admission is blocked in a cooperative Open.
		for running > 0 {
			err := <-finished
			running--
			if err != nil {
				failures = append(failures, err)
				stopping = true
			}
		}
		for _, lease := range leases {
			if err := lease.Close(); err != nil {
				failures = append(failures, err)
				stopping = true
				stop()
			}
		}
		if stopping || inputClosed || runCtx.Err() != nil {
			break
		}
	}
	if ctx.Err() != nil {
		failures = append(failures, ctx.Err())
	}
	return errors.Join(failures...)
}
