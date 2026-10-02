package sweeper

import "context"

func once(ctx context.Context, sweep func(context.Context) error) error {
	return sweep(ctx)
}
