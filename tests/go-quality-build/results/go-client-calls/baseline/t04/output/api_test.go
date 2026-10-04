package client_test

import (
	"context"
	"net/http"

	client "example.invalid/breaker-recovery"
)

var _ func(*http.Client, bool) *client.Client = client.New
var _ func(*client.Client, context.Context, string) ([]byte, error) = (*client.Client).Fetch
