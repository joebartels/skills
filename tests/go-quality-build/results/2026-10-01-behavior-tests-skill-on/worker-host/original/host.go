package sweeper

import (
	"context"
	"errors"
	"time"
)

func Run(ctx context.Context, interval time.Duration, sweep func(context.Context) error, release func() error) error {
	return errors.Join(once(ctx, sweep), release())
}
