package indexer

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidInterval = errors.New("invalid interval")

// Serve runs refresh synchronously and releases its resources after the last
// callback returns. Cancellation alone succeeds; independent failures survive it.
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
			if ctx.Err() != nil {
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

// A joined cancellation and independent error must retain the independent cause.
func onlyCancellation(err, cancellation error) bool {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		causes := joined.Unwrap()
		if len(causes) == 0 {
			return false
		}
		for _, cause := range causes {
			if !onlyCancellation(cause, cancellation) {
				return false
			}
		}
		return true
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		if cause := wrapped.Unwrap(); cause != nil {
			return onlyCancellation(cause, cancellation)
		}
	}
	return errors.Is(err, cancellation)
}
