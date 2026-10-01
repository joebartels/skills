package mirror

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// Refresh fetches and validates a remote snapshot before replacing path.
// On supported Unix filesystems, readers with an open old file retain it.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	if client == nil {
		return errors.New("refresh snapshot: nil HTTP client")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create snapshot request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("fetch snapshot: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch snapshot: HTTP status %d", response.StatusCode)
	}

	decoder := json.NewDecoder(response.Body)
	var items []Item
	if err := decoder.Decode(&items); err != nil {
		return fmt.Errorf("decode snapshot: %w", err)
	}
	if items == nil {
		return errors.New("decode snapshot: expected a JSON array")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return fmt.Errorf("decode snapshot trailing data: %w", err)
		}
		return errors.New("decode snapshot: trailing JSON value")
	}
	for i, item := range items {
		if strings.TrimSpace(item.Code) == "" || strings.TrimSpace(item.Label) == "" {
			return fmt.Errorf("validate snapshot: item %d has a blank code or label", i)
		}
	}
	data, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	return replaceSnapshot(path, data)
}

func replaceSnapshot(path string, data []byte) error {
	// A sibling temporary file allows replacement within the same filesystem.
	temporary, err := os.CreateTemp(filepath.Dir(path), ".snapshot-*")
	if err != nil {
		return fmt.Errorf("create snapshot temporary file: %w", err)
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write snapshot temporary file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close snapshot temporary file: %w", err)
	}
	if err := os.Rename(temporary.Name(), path); err != nil {
		return fmt.Errorf("replace snapshot: %w", err)
	}
	return nil
}
