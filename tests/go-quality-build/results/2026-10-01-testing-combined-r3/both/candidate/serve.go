package indexer

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidInterval = errors.New("invalid interval")

// Serve runs refresh immediately and sequentially, waiting interval after each
// successful completion. A valid invocation releases its resources exactly
// once after all callbacks finish. Parent cancellation alone is successful;
// independent callback failures and release failures retain their identities.
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
			if ctx.Err() != nil && onlyCancellation(err, ctx.Err()) {
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

// An errors.Join may contain both cancellation and an independent failure.
// Suppress ordinary parent cancellation only when every error leaf represents
// it; error values themselves need not be comparable.
func onlyCancellation(err, canceled error) bool {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		causes := joined.Unwrap()
		if len(causes) == 0 {
			return false
		}
		for _, cause := range causes {
			if cause == nil || !onlyCancellation(cause, canceled) {
				return false
			}
		}
		return true
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		if cause := wrapped.Unwrap(); cause != nil {
			return onlyCancellation(cause, canceled)
		}
	}
	return errors.Is(err, canceled)
}
