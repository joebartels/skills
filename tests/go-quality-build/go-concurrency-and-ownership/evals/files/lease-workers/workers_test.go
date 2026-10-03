package leaseworkers

import (
	"context"
	"testing"
)

type basicLease struct{ ran, closed bool }

func (l *basicLease) Run(context.Context, Job) error { l.ran = true; return nil }
func (l *basicLease) Close() error                   { l.closed = true; return nil }
func TestOneJob(t *testing.T) {
	jobs := make(chan Job, 1)
	jobs <- 1
	close(jobs)
	lease := new(basicLease)
	err := Serve(context.Background(), jobs, 1, func(context.Context, Job) (Lease, error) { return lease, nil })
	if err != nil || !lease.ran || !lease.closed {
		t.Fatalf("err=%v lease=%+v", err, lease)
	}
}
