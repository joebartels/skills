package indexer

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidInterval = errors.New("invalid interval")

// Serve runs refresh sequentially and releases its resources after it stops.
// Cancellation requests cooperative stopping and waits for callback cleanup.
func Serve(ctx context.Context, interval time.Duration, refresh func(context.Context) error, release func() error) (err error) {
	if interval <= 0 {
		return ErrInvalidInterval
	}
	defer func() { err = errors.Join(err, release()) }()
	for {
		if ctx.Err() != nil {
			return nil
		}
		if err := refresh(ctx); err != nil {
			if ctx.Err() != nil && onlyCancellation(err, ctx.Err(), context.Cause(ctx)) {
				return nil
			}
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

// Suppress ordinary context cancellation, while retaining independent errors
// even if a callback joins them with a cancellation error.
func onlyCancellation(err, cancellation, cause error) bool {
	if err == cancellation || err == cause {
		return true
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		children := joined.Unwrap()
		if len(children) == 0 {
			return false
		}
		for _, child := range children {
			if !onlyCancellation(child, cancellation, cause) {
				return false
			}
		}
		return true
	}
	if inner := errors.Unwrap(err); inner != nil {
		return onlyCancellation(inner, cancellation, cause)
	}
	return false
}
