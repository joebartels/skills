package indexer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Record struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

// Refresh fetches one JSON array and atomically publishes its validated records.
// It borrows client; acquired response bodies are closed before returning.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create refresh request: %w", err)
	}
	// A redirect is a rejected response, not permission for another GET. Copy
	// the supplied client to retain its transport, timeout and cookie policy
	// without changing the caller's redirect policy.
	oneRequest := *client
	oneRequest.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	response, err := oneRequest.Do(request)
	if err != nil {
		return fmt.Errorf("fetch records: %w", err)
	}
	// mutation: omit acquired body close
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch records: status %d; want 200", response.StatusCode)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read records: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	var records []Record
	if err := decoder.Decode(&records); err != nil {
		return fmt.Errorf("decode records: %w", err)
	}
	if records == nil {
		return errors.New("decode records: expected JSON array")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err != nil {
			return fmt.Errorf("decode trailing data: %w", err)
		}
		return errors.New("decode records: trailing JSON value")
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
	if err := writeAtomic(path, append(data, '\n')); err != nil {
		return fmt.Errorf("publish records: %w", err)
	}
	return nil
}
