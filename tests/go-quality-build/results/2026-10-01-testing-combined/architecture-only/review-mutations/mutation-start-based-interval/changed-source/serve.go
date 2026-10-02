package indexer

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidInterval = errors.New("invalid interval")

// Serve calls refresh immediately, then waits interval after each successful
// completion. Callbacks run sequentially with ctx. For a valid interval, release
// runs once after all callbacks return, including their cancellation cleanup.
// Parent cancellation alone (including a callback returning ctx.Err()) succeeds;
// other callback errors and release errors retain their identities.
func Serve(ctx context.Context, interval time.Duration, refresh func(context.Context) error, release func() error) (err error) {
	if interval <= 0 {
		return ErrInvalidInterval
	}
	defer func() { err = errors.Join(err, release()) }()
	for ctx.Err() == nil {
		started := time.Now()
		if err := refresh(ctx); err != nil {
			if cancellationOnly(err, ctx.Err(), context.Cause(ctx)) {
				return nil
			}
			return err
		}
		timer := time.NewTimer(interval - time.Since(started))
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
	}
	return nil
}

// A wrapped cancellation remains ordinary cancellation. A joined independent
// failure must still be returned, even if cancellation is another cause.
func cancellationOnly(err, canceled, cause error) bool {
	if canceled == nil {
		return false
	}
	if err == canceled || err == cause {
		return true
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		causes := joined.Unwrap()
		if len(causes) == 0 {
			return false
		}
		for _, child := range causes {
			if !cancellationOnly(child, canceled, cause) {
				return false
			}
		}
		return true
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		return cancellationOnly(wrapped.Unwrap(), canceled, cause)
	}
	return false
}
