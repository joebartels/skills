package indexer

import (
	"context"
	"errors"
	"time"
)

var ErrInvalidInterval = errors.New("invalid interval")

func Serve(ctx context.Context, interval time.Duration, refresh func(context.Context) error, release func() error) error {
	return ErrNotImplemented
}
