package workers

import (
	"context"
	"sync/atomic"
	"testing"
)

type basicLease struct {
	accepted *atomic.Int64
	closed   *atomic.Int64
}

func (l *basicLease) Run(context.Context, Job) error { l.accepted.Add(1); return nil }
func (l *basicLease) Close() error                   { l.closed.Add(1); return nil }
func TestServeFiniteJobs(t *testing.T) {
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	var accepted, closed atomic.Int64
	err := Serve(context.Background(), jobs, 2, func(context.Context, Job) (Lease, error) { return &basicLease{&accepted, &closed}, nil })
	if err != nil || accepted.Load() != 2 || closed.Load() != 2 {
		t.Fatalf("accepted=%d closed=%d err=%v", accepted.Load(), closed.Load(), err)
	}
}
