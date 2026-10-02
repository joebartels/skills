package poller

import (
	"context"
	"errors"
	"time"
)

// Poll runs sequential completion-relative cycles until cancellation or failure.
func Poll(ctx context.Context, interval time.Duration, callback func(context.Context) error) error {
	if interval <= 0 {
		return errors.New("poll interval must be positive")
	}
	for {
		if ctx.Err() != nil {
			return nil
		}
		callbackDone := make(chan error, 1)
		go func() { callbackDone <- callback(ctx) }()
		select {
		case <-ctx.Done(): return nil
		case err := <-callbackDone:
			if err != nil { return err }
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
