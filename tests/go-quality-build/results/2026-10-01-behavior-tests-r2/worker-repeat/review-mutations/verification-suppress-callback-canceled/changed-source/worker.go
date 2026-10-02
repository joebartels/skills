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
			if ctx.Err() != nil { return nil }
			return err
		}

		// Start the interval after the callback finishes, so slow callbacks do
		// not shorten the wait or overlap a later cycle.
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}
