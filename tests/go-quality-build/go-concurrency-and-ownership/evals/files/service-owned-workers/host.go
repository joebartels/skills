package workers

import (
	"context"
	"sync"
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
	var runs sync.WaitGroup
	for job := range jobs {
		lease, err := open(ctx, job)
		if err != nil {
			return err
		}
		runs.Add(1)
		go func(l Lease, j Job) { defer runs.Done(); _ = l.Run(ctx, j) }(lease, job)
		_ = lease.Close()
	}
	runs.Wait()
	return nil
}
