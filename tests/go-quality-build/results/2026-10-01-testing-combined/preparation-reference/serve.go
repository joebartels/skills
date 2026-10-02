package indexer

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidInterval = errors.New("invalid interval")

func Serve(ctx context.Context, interval time.Duration, refresh func(context.Context) error, release func() error) (out error) {
	if interval <= 0 {
		return ErrInvalidInterval
	}
	defer func() { out = errors.Join(out, release()) }()
	for {
		if ctx.Err() != nil {
			return nil
		}
		if err := refresh(ctx); err != nil {
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
