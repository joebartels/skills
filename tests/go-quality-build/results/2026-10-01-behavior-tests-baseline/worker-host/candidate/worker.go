package sweeper

import (
	"context"
	"time"
)

func work(ctx context.Context, interval time.Duration, sweep func(context.Context) error) error {
	for {
		if ctx.Err() != nil {
			return nil
		}
		if err := sweep(ctx); err != nil {
			return err
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
}
