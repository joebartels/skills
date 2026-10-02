package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Record struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

// Refresh fetches and validates a complete array before atomically publishing it.
// The supplied client remains owned by the caller.
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
	decoder := json.NewDecoder(response.Body)
	var records []Record
	if err := decoder.Decode(&records); err != nil {
		return fmt.Errorf("decode records: %w", err)
	}
	if records == nil {
		return fmt.Errorf("decode records: expected a JSON array")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode records: trailing JSON value")
		}
		return fmt.Errorf("decode records: trailing data: %w", err)
	}
	for n, record := range records {
		if !keyPattern.MatchString(record.Key) || strings.TrimSpace(record.Text) == "" {
			return fmt.Errorf("record %d: %w", n+1, ErrInvalidRecord)
		}
	}
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("encode records: %w", err)
	}
	data = append(data, '\n')
	if err := publish(path, data); err != nil {
		return fmt.Errorf("publish records: %w", err)
	}
	return nil
}

func publish(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return nil
}
