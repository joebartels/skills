package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Metadata struct {
	Name     string `json:"name"`
	Revision int    `json:"revision"`
}

// Fetch reads complete validated metadata from endpoint.
func Fetch(ctx context.Context, client *http.Client, endpoint string) (Metadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Metadata{}, err
	}
	response, err := client.Do(req)
	if err != nil {
		return Metadata{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Metadata{}, fmt.Errorf("metadata HTTP status %d", response.StatusCode)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return Metadata{}, err
	}
	var result Metadata
	if err := json.NewDecoder(strings.NewReader(string(data))).Decode(&result); err != nil {
		return Metadata{}, err
	}
	if strings.TrimSpace(result.Name) == "" || result.Revision <= 0 {
		return Metadata{}, fmt.Errorf("invalid metadata")
	}
	return result, nil
}
