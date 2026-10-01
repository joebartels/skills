package sweeper

import (
	"context"
	"errors"
	"time"
)

// Run owns a periodic sweep until ctx is canceled or sweep returns an error.
// The first sweep is immediate, and each later sweep waits interval after the
// previous callback completes. Callbacks never overlap.
//
// Cancellation stops future sweeps and waits for an active callback to finish.
// A callback error is preserved even if ctx was canceled during that callback.
// Run cancels and joins the worker before calling release exactly once, and
// joins its error with the callback error. A nonpositive interval is rejected
// without invoking either callback.
func Run(ctx context.Context, interval time.Duration, sweep func(context.Context) error, release func() error) error {
	if interval <= 0 {
		return errors.New("sweeper: interval must be positive")
	}

	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() {
		done <- work(workerCtx, interval, sweep)
	}()

	var workerErr error
	select {
	case workerErr = <-done:
		cancel()
	case <-ctx.Done():
		cancel()
		workerErr = <-done
	}
	return errors.Join(workerErr, release())
}
