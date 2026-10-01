package sweeper

import (
	"context"
	"errors"
	"time"
)

// Run owns a periodic worker until ctx is canceled or sweep fails. The first
// sweep starts immediately; subsequent sweeps start interval after completion.
// Callback errors take precedence over concurrent cancellation.
//
// Run cancels the worker context and waits for all callback cleanup before
// calling release exactly once. It joins worker and release errors. An invalid
// interval returns an error without invoking sweep or release.
func Run(ctx context.Context, interval time.Duration, sweep func(context.Context) error, release func() error) error {
	if interval <= 0 {
		return errors.New("sweep interval must be positive")
	}

	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() {
		done <- periodic(workerCtx, interval, sweep)
	}()

	err := <-done
	cancel()
	return errors.Join(err, release())
}
