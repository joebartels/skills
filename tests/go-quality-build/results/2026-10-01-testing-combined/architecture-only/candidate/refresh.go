package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Record is a key/text pair in the refresh wire and file formats.
type Record struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

// Refresh fetches one array of records and atomically publishes compact JSON
// followed by a newline. Rejection leaves an existing destination unchanged.
// The client is borrowed; Refresh closes the response body it acquires.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create refresh request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch records: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch records: HTTP status %d", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	var records []Record
	if err := decoder.Decode(&records); err != nil {
		return fmt.Errorf("decode records: %w", err)
	}
	if records == nil {
		return fmt.Errorf("decode records: expected a JSON array")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return fmt.Errorf("read trailing data: %w", err)
		}
		return fmt.Errorf("decode records: unexpected trailing JSON")
	}
	for n, record := range records {
		if err := validateRecord(record.Key, record.Text); err != nil {
			return fmt.Errorf("record %d: %w", n+1, err)
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
