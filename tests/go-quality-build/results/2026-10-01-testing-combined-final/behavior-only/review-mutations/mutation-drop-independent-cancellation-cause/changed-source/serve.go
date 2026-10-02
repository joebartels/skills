package indexer

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidInterval = errors.New("invalid interval")

// Serve runs refresh sequentially and releases its resources after all work ends.
// Ordinary context cancellation is successful shutdown; independent errors survive.
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
			if ctx.Err() != nil && errors.Is(err, ctx.Err()) {
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

// Inspect joined causes individually so cancellation cannot hide another failure.
func cancellationOnly(err, cancellation error) bool {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		causes := joined.Unwrap()
		if len(causes) == 0 {
			return false
		}
		for _, cause := range causes {
			if !cancellationOnly(cause, cancellation) {
				return false
			}
		}
		return true
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		return cancellationOnly(wrapped.Unwrap(), cancellation)
	}
	return errors.Is(err, cancellation)
}
