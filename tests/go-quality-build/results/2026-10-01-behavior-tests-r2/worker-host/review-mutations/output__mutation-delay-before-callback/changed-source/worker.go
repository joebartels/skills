package sweeper

import (
	"context"
	"time"
)

func runWorker(ctx context.Context, interval time.Duration, sweep func(context.Context) error) error {
	for {
		if ctx.Err() != nil {
			return nil
		}
		timer := time.NewTimer(interval)
		if err := sweep(ctx); err != nil {
			return err
		}

		// Start the delay after the callback returns, rather than accumulating
		// ticks while it runs. Only this goroutine invokes the callback.
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}
