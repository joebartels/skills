package indexer

import (
	"context"
	"net/http"
)

type Record struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	return ErrNotImplemented
}
