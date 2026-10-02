package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Record struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

// Refresh fetches and validates one snapshot before atomically publishing it.
// The caller owns client; Refresh closes every response body it acquires.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create refresh request: %w", err)
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("fetch records: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch records: HTTP status %d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read records: %w", err)
	}
	var records []Record
	if err := json.Unmarshal(body, &records); err != nil {
		return fmt.Errorf("decode records: %w", err)
	}
	if records == nil {
		return fmt.Errorf("decode records: expected JSON array")
	}
	for n, record := range records {
		if !validRecord(record.Key, record.Text) {
			return fmt.Errorf("record %d: %w", n+1, ErrInvalidRecord)
		}
	}
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("encode records: %w", err)
	}
	if err := publish(path, append(data, '\n')); err != nil {
		return fmt.Errorf("publish records: %w", err)
	}
	return nil
}

func publish(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".snapshot-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
