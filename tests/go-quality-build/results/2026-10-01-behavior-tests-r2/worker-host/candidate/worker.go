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
		if err := sweep(ctx); err != nil {
			return err
		}

		// Start the delay after the callback returns, rather than accumulating
		// ticks while it runs. Only this goroutine invokes the callback.
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}
