package indexer

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidInterval = errors.New("invalid interval")

// Serve invokes refresh sequentially and waits interval after each completion.
// It releases resources once all callbacks finish. Callback errors, including
// those returned during cancellation cleanup, remain visible to the caller.
func Serve(ctx context.Context, interval time.Duration, refresh func(context.Context) error, release func() error) (err error) {
	if interval <= 0 {
		return ErrInvalidInterval
	}
	defer func() { err = errors.Join(err, release()) }()
	for {
		if ctx.Err() != nil {
			return nil
		}
		callbackDone := make(chan error, 1)
		go func() { callbackDone <- refresh(ctx) }()
		select {
		case callbackErr := <-callbackDone:
			if callbackErr != nil { return callbackErr }
		case <-ctx.Done():
			return nil
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
