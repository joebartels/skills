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

// Refresh fetches and validates a complete snapshot before replacing path.
// The supplied client is borrowed; every acquired response body is closed.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create refresh request: %w", err)
	}
	if client == nil {
		return fmt.Errorf("refresh requires an HTTP client")
	}
	// A redirect would issue another GET and accept a non-200 initial response.
	// Copy the client so the caller's redirect policy remains unchanged.
	oneRequest := *client
	oneRequest.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	response, err := oneRequest.Do(request)
	if err != nil {
		return fmt.Errorf("fetch snapshot: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch snapshot: HTTP status %d", response.StatusCode)
	}
	decoder := json.NewDecoder(response.Body)
	var records []Record
	if err := decoder.Decode(&records); err != nil {
		return fmt.Errorf("decode snapshot: %w", err)
	}
	if records == nil {
		return fmt.Errorf("decode snapshot: expected a JSON array")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode snapshot: trailing JSON value")
		}
		return fmt.Errorf("decode snapshot trailing data: %w", err)
	}
	for n, record := range records {
		if err := validateRecord(record.Key, record.Text); err != nil {
			return fmt.Errorf("snapshot record %d: %w", n+1, err)
		}
	}
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	data = append(data, '\n')
	if err := writeAtomic(path, data); err != nil {
		return fmt.Errorf("publish snapshot: %w", err)
	}
	return nil
}
