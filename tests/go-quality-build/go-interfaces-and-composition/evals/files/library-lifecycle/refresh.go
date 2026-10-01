package refreshkit

import (
	"context"
	"errors"
)

type Refresher struct{ refresh func(context.Context) error }

func New(refresh func(context.Context) error) (*Refresher, error) {
	if refresh == nil {
		return nil, errors.New("nil refresh callback")
	}
	return &Refresher{refresh: refresh}, nil
}
func (r *Refresher) Refresh(ctx context.Context) error { return r.refresh(ctx) }
