package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Record struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

func Refresh(ctx context.Context, client *http.Client, endpoint, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	var records []Record
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&records); err != nil {
		return err
	}
	if records == nil {
		return ErrInvalidRecord
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return fmt.Errorf("trailing data: %v", err)
	}
	for _, r := range records {
		if !keyPattern.MatchString(r.Key) || strings.TrimSpace(r.Text) == "" {
			return ErrInvalidRecord
		}
	}
	b, err := json.Marshal(records)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	f, err := os.CreateTemp(filepath.Dir(path), ".snapshot-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
