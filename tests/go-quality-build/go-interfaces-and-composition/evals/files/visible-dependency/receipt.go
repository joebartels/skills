package receipts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Sender struct{}

func (Sender) Send(ctx context.Context, id string) error {
	body, err := json.Marshal(struct {
		ID string `json:"id"`
	}{id})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, os.Getenv("RECEIPT_URL"), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("receipt status: %d", response.StatusCode)
	}
	return nil
}
