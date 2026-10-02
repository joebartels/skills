package sweeper

import (
	"context"
	"errors"
	"time"
)

// Run owns a periodic sweep until ctx is canceled or a callback fails. Sweeps
// start immediately and then wait interval after each successful completion.
// It cancels and joins the worker before releasing resources and returning.
// Callback errors, including errors returned during cancellation, are preserved
// and joined with any release error. A callback must finish after cancellation
// for Run to return. A nonpositive interval is rejected without calling either
// callback.
func Run(ctx context.Context, interval time.Duration, sweep func(context.Context) error, release func() error) error {
	if interval <= 0 {
		return errors.New("sweep interval must be positive")
	}

	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() {
		done <- runWorker(workerCtx, interval, sweep)
	}()

	var workerErr error
	select {
	case workerErr = <-done:
		cancel()
	case <-ctx.Done():
		cancel()
		releaseErr := release()
		workerErr = <-done
		return errors.Join(workerErr, releaseErr)
	}
	return errors.Join(workerErr, release())
}
