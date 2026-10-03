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

type runResult struct {
	index int
	err   error
}

type ownedLease struct {
	lease   Lease
	job     Job
	started bool
}

// Serve runs supplied jobs using leases acquired by the host.
func Serve(ctx context.Context, jobs <-chan Job, limit int, open func(context.Context, Job) (Lease, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var errs []error
	var cohort []ownedLease
	runs := make(chan runResult, limit)
	running := 0
	input := jobs
	stop := false

	startRun := func(i int) {
		cohort[i].started = true
		running++
		l, j := cohort[i].lease, cohort[i].job
		go func() { runs <- runResult{index: i, err: l.Run(workCtx, j)} }()
	}
	joinRuns := func() {
		for running > 0 {
			r := <-runs
			running--
			if r.err != nil {
				errs = append(errs, r.err)
				cancel()
				stop = true
			}
		}
	}
	closeCohort := func() {
		for i := range cohort {
			if err := cohort[i].lease.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		cohort = nil
	}

	for !stop {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			cancel()
			stop = true
			break
		}
		if len(cohort) == limit {
			// Capacity remains occupied until every Run in this cohort has ended
			// and every lease has completed Close.
			joinRuns()
			if ctx.Err() != nil {
				errs = append(errs, ctx.Err())
				stop = true
				break
			}
			closeCohort()
			continue
		}
		if input == nil {
			if running == 0 {
				break
			}
			select {
			case r := <-runs:
				running--
				if r.err != nil {
					errs = append(errs, r.err)
					cancel()
					stop = true
				}
			case <-ctx.Done():
				errs = append(errs, ctx.Err())
				cancel()
				stop = true
			}
			continue
		}
		// Opening asynchronously lets cancellation or an already-running job
		// failure interrupt admission. The result is always collected so a lease
		// returned concurrently with stop is still released.
		jobCh := make(chan Job, 1)
		select {
		case j, ok := <-input:
			if !ok {
				input = nil
				continue
			}
			jobCh <- j
		case r := <-runs:
			running--
			if r.err != nil {
				errs = append(errs, r.err)
				cancel()
				stop = true
			}
			continue
		case <-ctx.Done():
			errs = append(errs, ctx.Err())
			cancel()
			stop = true
			continue
		}
		j := <-jobCh
		type openResult struct {
			lease Lease
			err   error
		}
		opened := make(chan openResult, 1)
		go func() { l, err := open(workCtx, j); opened <- openResult{l, err} }()
		var result openResult
		waiting := true
		for waiting {
			select {
			case result = <-opened:
				waiting = false
			case r := <-runs:
				running--
				if r.err != nil {
					errs = append(errs, r.err)
					cancel()
					stop = true
				}
			case <-ctx.Done():
				errs = append(errs, ctx.Err())
				cancel()
				stop = true
			}
			if stop && waiting {
				// Open is contractually cooperative with workCtx.
				result = <-opened
				waiting = false
			}
		}
		if result.lease != nil {
			cohort = append(cohort, ownedLease{lease: result.lease, job: j})
		}
		if result.err == nil && result.lease == nil {
			result.err = errors.New("open returned a nil lease without an error")
		}
		if result.err != nil {
			errs = append(errs, result.err)
			cancel()
			stop = true
		}
		if result.lease != nil && !stop {
			startRun(len(cohort) - 1)
		}
	}
	if stop {
		cancel()
	}
	joinRuns()
	closeCohort()
	return errors.Join(errs...)
}
