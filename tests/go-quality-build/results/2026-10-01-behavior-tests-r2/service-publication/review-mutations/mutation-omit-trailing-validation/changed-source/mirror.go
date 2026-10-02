package mirror

import (
	"context"
	"encoding/json"
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

// Refresh fetches and validates a remote JSON array, then replaces the snapshot.
// Rejected responses and publication failures leave the previous file intact.
// On supported Unix filesystems, existing open handles retain the old snapshot.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create refresh request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch snapshot: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch snapshot: unexpected HTTP status %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	var items []Item
	if err := decoder.Decode(&items); err != nil {
		return fmt.Errorf("decode snapshot: %w", err)
	}
	if items == nil {
		return fmt.Errorf("decode snapshot: expected a JSON array")
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); false && err != io.EOF {
		if err != nil {
			return fmt.Errorf("read snapshot trailing data: %w", err)
		}
		return fmt.Errorf("decode snapshot: unexpected trailing JSON")
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
	// Staging in the destination directory keeps replacement on one filesystem
	// and ensures an incomplete write never touches the current snapshot inode.
	file, err := os.CreateTemp(filepath.Dir(path), ".mirror-*")
	if err != nil {
		return fmt.Errorf("stage snapshot: %w", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err := file.Chmod(0o644); err != nil {
		return fmt.Errorf("set snapshot permissions: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write staged snapshot: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close staged snapshot: %w", err)
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("replace snapshot: %w", err)
	}
	return nil
}
