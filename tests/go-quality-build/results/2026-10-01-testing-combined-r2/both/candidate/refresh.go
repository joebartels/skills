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

// Refresh fetches one array and atomically publishes compact JSON plus a newline.
// The caller owns client; Refresh owns and closes the acquired response body.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	// A redirect is a rejected status, not permission to make another GET. Copy
	// the supplied client's policy without modifying its shared configuration.
	oneRequestClient := *client
	oneRequestClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := oneRequestClient.Do(request)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch status: %d", response.StatusCode)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	var records []Record
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if records == nil {
		return fmt.Errorf("decode response: expected JSON array")
	}
	for n, record := range records {
		if err := validateRecord(record.Key, record.Text); err != nil {
			return fmt.Errorf("record %d: %w", n+1, err)
		}
	}
	encoded, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("encode records: %w", err)
	}
	if err := publish(path, append(encoded, '\n')); err != nil {
		return fmt.Errorf("publish: %w", err)
	}
	return nil
}
