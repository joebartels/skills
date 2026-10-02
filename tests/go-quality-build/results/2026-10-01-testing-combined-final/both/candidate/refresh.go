package indexer

import (
	"bytes"
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

// Refresh fetches and validates one array before atomically publishing compact JSON.
// The client is borrowed; the response body is owned and closed by this call.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("refresh request: %w", err)
	}
	// A redirect would make another GET and hide the initial non-200 status.
	requestClient := *client
	requestClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	response, err := requestClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh fetch: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh status: %d", response.StatusCode)
	}
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("refresh read: %w", err)
	}
	payload = bytes.TrimSpace(payload)
	if len(payload) == 0 || payload[0] != '[' {
		return fmt.Errorf("refresh decode: expected a JSON array")
	}
	var records []Record
	if err := json.Unmarshal(payload, &records); err != nil {
		return fmt.Errorf("refresh decode: %w", err)
	}
	for position, record := range records {
		if err := validateRecord(record.Key, record.Text); err != nil {
			return fmt.Errorf("refresh record %d: %w", position+1, err)
		}
	}
	encoded, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("refresh encode: %w", err)
	}
	if err := replaceFile(path, append(encoded, '\n')); err != nil {
		return fmt.Errorf("refresh publish: %w", err)
	}
	return nil
}
