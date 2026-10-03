package leaseworkers

import (
	"context"
	"errors"
)

type Job int

type Lease interface {
	Run(context.Context, Job) error
	Close() error
}

func Serve(ctx context.Context, jobs <-chan Job, limit int, open func(context.Context, Job) (Lease, error)) error {
	for job := range jobs {
		lease, err := open(ctx, job)
		if err != nil {
			return err
		}
		if err := errors.Join(lease.Run(ctx, job), lease.Close()); err != nil {
			return err
		}
	}
	return nil
}
