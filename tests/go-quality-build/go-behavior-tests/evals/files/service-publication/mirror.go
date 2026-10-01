package mirror

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

type Item struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

// Write writes the original local-only snapshot.
func Write(path string, items []Item) error {
	data, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Refresh is reserved for the next minor release.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	return errors.New("remote refresh not implemented")
}
