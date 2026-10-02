package indexer

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidInterval = errors.New("invalid interval")

// Serve runs refresh immediately and waits interval after each successful call.
// Calls are synchronous: release runs once only after the last call returns.
// A callback error caused solely by the parent context represents cancellation;
// other callback errors and release errors are retained.
func Serve(ctx context.Context, interval time.Duration, refresh func(context.Context) error, release func() error) (err error) {
	if interval <= 0 {
		return ErrInvalidInterval
	}
	defer func() { err = errors.Join(err, release()) }()
	for {
		if ctx.Err() != nil {
			return nil
		}
		result := make(chan error, 1)
		go func() { result <- refresh(ctx) }()
		select {
		case <-ctx.Done():
			return nil // MUTATION: release before callback cleanup returns
		case callbackErr := <-result:
			if callbackErr != nil {
				if cancellationOnly(callbackErr, ctx.Err(), context.Cause(ctx)) {
					return nil
				}
				return callbackErr
			}
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

// Do not discard an independent failure joined with a context error.
func cancellationOnly(err, parent, cause error) bool {
	if parent == nil {
		return false
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		causes := joined.Unwrap()
		if len(causes) == 0 {
			return false
		}
		for _, child := range causes {
			if !cancellationOnly(child, parent, cause) {
				return false
			}
		}
		return true
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		return cancellationOnly(wrapped.Unwrap(), parent, cause)
	}
	return errors.Is(err, parent) || errors.Is(err, cause)
}
