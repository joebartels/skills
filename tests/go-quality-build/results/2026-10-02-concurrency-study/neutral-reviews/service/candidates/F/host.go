package workers

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Job identifies a requested operation.
type Job int

// Lease supplies the existing cooperative operation and release boundary.
type Lease interface {
	Run(context.Context, Job) error
	Close() error
}

type result struct {
	job   Job
	lease Lease
	err   error
	open  bool
}

// Serve runs supplied jobs using leases acquired by the host. Capacity is
// occupied through Close, and every cohort Run joins before any Close begins.
func Serve(ctx context.Context, jobs <-chan Job, limit int, open func(context.Context, Job) (Lease, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var errs []error
	for {
		workCtx, stop := context.WithCancel(ctx)
		events := make(chan result, limit*2)
		var workers sync.WaitGroup
		var leases []result
		active := 0
		inputClosed, failed := false, false
		for !inputClosed && !failed && active < limit {
			if err := ctx.Err(); err != nil {
				errs = append(errs, err)
				failed = true
				stop()
				break
			}
			select {
			case <-ctx.Done():
				errs = append(errs, ctx.Err())
				failed = true
				stop()
			case ev := <-events:
				if ev.open {
					leases = append(leases, ev)
				}
				if ev.err != nil {
					errs = append(errs, ev.err)
					failed = true
					stop()
				}
				if !ev.open {
					active--
				}
			case job, ok := <-jobs:
				if !ok {
					inputClosed = true
					break
				}
				active++
				workers.Add(1)
				go func(job Job) {
					defer workers.Done()
					lease, err := open(workCtx, job)
					if err != nil {
						events <- result{job: job, err: fmt.Errorf("open job %d: %w", job, err)}
						stop()
						return
					}
					if lease == nil {
						events <- result{job: job, err: fmt.Errorf("open job %d returned a nil lease", job)}
						stop()
						return
					}
					events <- result{job: job, lease: lease, open: true}
					err = lease.Run(workCtx, job)
					events <- result{job: job, err: wrap("run", job, err)}
					if err != nil {
						stop()
					}
				}(job)
			}
		}
		stop()
		workers.Wait()
		for len(events) > 0 {
			ev := <-events
			if ev.open {
				leases = append(leases, ev)
			} else if ev.err != nil && !hasError(errs, ev.err) {
				errs = append(errs, ev.err)
			}
		}
		for _, item := range leases {
			if err := item.lease.Close(); err != nil {
				errs = append(errs, fmt.Errorf("close job %d: %w", item.job, err))
			}
		}
		stop()
		if err := ctx.Err(); err != nil && !hasError(errs, err) {
			errs = append(errs, err)
		}
		if len(errs) > 0 || ctx.Err() != nil {
			return errors.Join(errs...)
		}
		if inputClosed {
			return nil
		}
	}
}

func wrap(stage string, job Job, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s job %d: %w", stage, job, err)
}
func hasError(errs []error, target error) bool {
	for _, err := range errs {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}
