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
// On supported Unix filesystems, existing open handles retain the old snapshot.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create snapshot request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch snapshot: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch snapshot: HTTP status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read snapshot: %w", err)
	}
	var items []Item
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("decode snapshot: %w", err)
	}
	if items == nil {
		return errors.New("snapshot must be a JSON array")
	}
	for i, item := range items {
		if strings.TrimSpace(item.Code) == "" || strings.TrimSpace(item.Label) == "" {
			return fmt.Errorf("snapshot item %d has a blank code or label", i)
		}
	}
	data, err = json.Marshal(items)
	if err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	// Keep the temporary file in the destination directory so publication is
	// a single filesystem rename, never a truncation of the previous file.
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create snapshot file: %w", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err := file.Chmod(0o644); err != nil {
		return fmt.Errorf("set snapshot permissions: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close snapshot file: %w", err)
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("publish snapshot: %w", err)
	}
	return nil
}
