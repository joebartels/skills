package workers

import (
	"context"
	"errors"
	"sync"
)

type Job int
type Lease interface {
	Run(context.Context, Job) error
	Close() error
}

func Serve(ctx context.Context, jobs <-chan Job, limit int, open func(context.Context, Job) (Lease, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	var runs sync.WaitGroup
	var mu sync.Mutex
	var failures []error
	collect := func(err error) {
		if err == nil {
			return
		}
		mu.Lock()
		failures = append(failures, err)
		mu.Unlock()
		cancel()
	}
	var cohort []Lease
	finish := func() {
		runs.Wait()
		for _, lease := range cohort {
			collect(lease.Close())
		}
		cohort = nil
	}
	stopping := false
admission:
	for {
		if len(cohort) == limit {
			finish()
			if child.Err() != nil {
				stopping = true
				break
			}
		}
		var job Job
		var ok bool
		select {
		case <-child.Done():
			stopping = true
			break admission
		case job, ok = <-jobs:
		}
		if !ok {
			break
		}
		if child.Err() != nil {
			stopping = true
			break
		}
		lease, err := open(child, job)
		if err != nil {
			collect(err)
			break
		}
		cohort = append(cohort, lease)
		if child.Err() != nil {
			stopping = true
			break
		}
		runs.Add(1)
		go func(l Lease, j Job) { defer runs.Done(); collect(l.Run(child, j)) }(lease, job)
	}
	finish()
	if stopping {
		collect(ctx.Err())
	}
	mu.Lock()
	defer mu.Unlock()
	return errors.Join(failures...)
}
