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

// Refresh fetches and validates one record array before publishing its snapshot.
// The caller owns client; Refresh closes each acquired response body.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	// Following redirects would issue extra GETs and accept a non-200 response.
	requestClient := *client
	requestClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := requestClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch records: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch records: status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read records: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	var records []Record
	if err := decoder.Decode(&records); err != nil {
		return fmt.Errorf("decode records: %w", err)
	}
	if records == nil {
		return fmt.Errorf("decode records: expected an array")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("decode records: trailing data")
	}
	for n, record := range records {
		if err := validateRecord(record.Key, record.Text); err != nil {
			return fmt.Errorf("record %d: %w", n+1, err)
		}
	}
	data, err = json.Marshal(records)
	if err != nil {
		return fmt.Errorf("encode records: %w", err)
	}
	if err := publish(path, append(data, '\n')); err != nil {
		return fmt.Errorf("publish records: %w", err)
	}
	return nil
}
