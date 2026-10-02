package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Record struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

// Refresh fetches and validates a snapshot before atomically publishing it.
// The client is borrowed; its transport and other policies remain caller-owned.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	if client == nil {
		return fmt.Errorf("refresh: nil HTTP client")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("refresh request: %w", err)
	}
	// Following redirects would make additional requests and could turn a
	// rejected status into an accepted one. Copy the client, preserving the
	// caller's transport and timeout without changing its redirect policy.
	oneRequest := *client
	oneRequest.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	response, err := oneRequest.Do(request)
	if err != nil {
		return fmt.Errorf("refresh fetch: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh status: %s", response.Status)
	}
	decoder := json.NewDecoder(response.Body)
	var records []Record
	if err := decoder.Decode(&records); err != nil {
		return fmt.Errorf("refresh decode: %w", err)
	}
	if records == nil {
		return fmt.Errorf("refresh decode: expected JSON array")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("refresh decode: trailing JSON value")
		}
		return fmt.Errorf("refresh trailing data: %w", err)
	}
	for position, record := range records {
		if !validRecord(record.Key, record.Text) {
			return fmt.Errorf("refresh record %d: %w", position+1, ErrInvalidRecord)
		}
	}
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("refresh encode: %w", err)
	}
	if err := writeAtomic(path, append(data, '\n')); err != nil {
		return fmt.Errorf("refresh publish: %w", err)
	}
	return nil
}
