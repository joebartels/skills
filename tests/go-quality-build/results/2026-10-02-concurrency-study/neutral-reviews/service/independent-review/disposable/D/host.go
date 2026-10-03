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

// Serve runs supplied jobs using at most limit leases at a time.
func Serve(ctx context.Context, jobs <-chan Job, limit int, open func(context.Context, Job) (Lease, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	workCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var result error

	for {
		if err := ctx.Err(); err != nil {
			result = errors.Join(result, err)
			break
		}
		cohort := make([]Job, 0, limit)
		select {
		case <-ctx.Done():
			result = errors.Join(result, ctx.Err())
			goto stop
		case job, ok := <-jobs:
			if !ok {
				return result
			}
			cohort = append(cohort, job)
		}
		for len(cohort) < limit {
			select {
			case job, ok := <-jobs:
				if !ok {
					goto process
				}
				cohort = append(cohort, job)
			default:
				goto process
			}
		}
	process:
		if len(cohort) == 0 {
			return result
		}
		leases := make([]Lease, len(cohort))
		type openResult struct {
			i     int
			lease Lease
			err   error
		}
		opened := make(chan openResult, len(cohort))
		for i, job := range cohort {
			go func(i int, job Job) {
				lease, err := open(workCtx, job)
				opened <- openResult{i: i, lease: lease, err: err}
			}(i, job)
		}
		failed := false
		callerDone := ctx.Done()
		for received := 0; received < len(cohort); received++ {
			select {
			case outcome := <-opened:
				leases[outcome.i] = outcome.lease
				if outcome.err != nil {
					result = errors.Join(result, outcome.err)
					failed = true
					cancel()
				} else if outcome.lease == nil {
					result = errors.Join(result, errors.New("open returned a nil lease without error"))
					failed = true
					cancel()
				}
			case <-callerDone:
				result = errors.Join(result, ctx.Err())
				failed = true
				cancel()
				// Open cooperates with this stop; continue collecting owned leases.
				callerDone = nil
				received--
			}
		}
		if ctx.Err() != nil {
			result = errors.Join(result, ctx.Err())
			failed = true
		}
		type runResult struct{ err error }
		done := make(chan runResult, len(leases))
		var runs sync.WaitGroup
		runCount := 0
		for i, lease := range leases {
			if lease == nil {
				continue
			}
			if failed {
				continue
			}
			runs.Add(1)
			runCount++
			go func(i int, l Lease, j Job) {
				defer runs.Done()
				done <- runResult{err: l.Run(workCtx, j)}
			}(i, lease, cohort[i])
		}
		if !failed {
			// Observe failures while siblings are running so they can be asked to stop.
			completed := 0
			for completed < runCount {
				select {
				case outcome := <-done:
					completed++
					if outcome.err != nil {
						result = errors.Join(result, outcome.err)
						failed = true
						cancel()
					}
				case <-ctx.Done():
					result = errors.Join(result, ctx.Err())
					failed = true
					cancel()
					// The caller's done channel stays ready; drain remaining runs below.
					goto joinRuns
				}
			}
		}
	joinRuns:
		runs.Wait()
		// Collect any outcomes not consumed by the live failure observer.
		for {
			select {
			case outcome := <-done:
				if outcome.err != nil {
					result = errors.Join(result, outcome.err)
				}
			default:
				goto outcomesCollected
			}
		}
	outcomesCollected:
		if failed {
			cancel()
		}
		// The cohort-wide Run join above is the release barrier.
		for _, lease := range leases {
			if lease != nil {
				result = errors.Join(result, lease.Close())
			}
		}
		if failed || ctx.Err() != nil {
			if ctx.Err() != nil {
				result = errors.Join(result, ctx.Err())
			}
			break
		}
	}
	return result

stop:
	cancel()
	return result
}
