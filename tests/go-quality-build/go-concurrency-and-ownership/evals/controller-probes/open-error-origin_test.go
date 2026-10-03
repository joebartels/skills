package leaseworkers

import (
	"context"
	"errors"
	"testing"
	"time"
)

// Post-inspection diagnostic, not a frozen comparison probe. The original
// README declares failed Open results independent. This Open has an independent
// failure before peer stopping, but finishes its cleanup only after that stop.
// Its bare standard sentinel must not be reclassified from return-time state.
func TestDiagnosticIndependentOpenErrorBeforePeerStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan Job, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	openFailed := make(chan struct{})
	peerFailure := errors.New("independent peer Run failure")
	result := make(chan error, 1)
	finished := make(chan struct{})
	t.Cleanup(func() {
		cancel()
		select {
		case <-finished:
		case <-time.After(3 * time.Second):
			t.Error("diagnostic Serve did not join after cancellation")
		}
	})
	go func() {
		defer close(finished)
		result <- Serve(ctx, jobs, 2, func(workerCtx context.Context, job Job) (Lease, error) {
			if job == 2 {
				// Capture an independent acquisition outcome before any peer stop.
				failure := context.Canceled
				close(openFailed)
				<-workerCtx.Done()
				return nil, failure
			}
			return openOriginLease{run: func() error {
				select {
				case <-openFailed:
					return peerFailure
				case <-workerCtx.Done():
					return workerCtx.Err()
				}
			}}, nil
		})
	}()
	select {
	case err := <-result:
		if !errors.Is(err, peerFailure) {
			t.Errorf("Serve = %v, missing independent peer failure", err)
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Serve = %v, missing independent Open failure established before peer stop", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("diagnostic Serve did not return")
	}
}

type openOriginLease struct{ run func() error }

func (l openOriginLease) Run(context.Context, Job) error { return l.run() }
func (openOriginLease) Close() error                     { return nil }
