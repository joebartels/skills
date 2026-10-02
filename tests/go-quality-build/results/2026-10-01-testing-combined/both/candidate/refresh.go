package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Record is the key/text representation used by refresh responses.
type Record struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

// Refresh fetches and validates one JSON array, then atomically publishes its
// compact representation plus a newline. The supplied client remains borrowed.
func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create refresh request: %w", err)
	}
	// A redirect would issue another GET and could hide a rejected status. Copy
	// the borrowed client's configuration without changing its redirect policy.
	requestClient := *client
	requestClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := requestClient.Do(request)
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
		return fmt.Errorf("expected a JSON array: %w", ErrInvalidRecord)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("unexpected data after records")
		}
		return fmt.Errorf("read after records: %w", err)
	}
	for n, record := range records {
		if err := validateRecord(record); err != nil {
			return fmt.Errorf("record %d: %w", n+1, err)
		}
	}
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("encode records: %w", err)
	}
	if err := replaceFile(path, append(data, '\n')); err != nil {
		return fmt.Errorf("publish records: %w", err)
	}
	return nil
}
