package sweeper

import (
	"context"
	"errors"
	"time"
)

// Run sweeps immediately and then waits interval after each successful sweep.
// Cancellation stops future sweeps and returns nil unless sweep returns an
// error. A sweep error takes precedence over concurrent cancellation.
// Run waits for sweep to finish, cancels its context, and invokes release once
// before returning, preserving both sweep and release errors.
// A nonpositive interval is rejected without invoking either callback.
func Run(ctx context.Context, interval time.Duration, sweep func(context.Context) error, release func() error) error {
	if interval <= 0 {
		return errors.New("sweep interval must be positive")
	}

	workerCtx, cancel := context.WithCancel(ctx)
	err := runWorker(workerCtx, interval, sweep)
	cancel()
	return errors.Join(err, release())
}
