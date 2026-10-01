package dispatch

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Service processes one delivery request at a time. Callers serialize access.
type Service struct {
	Dir, NotifyURL string
	Client         *http.Client
}

// Submit validates an ID, writes its record, then notifies the external service.
// A notification failure leaves the record in place and returns an error; no retry is promised.
func (s *Service) Submit(ctx context.Context, id string) error {
	if id == "" || strings.ContainsAny(id, "/\\") || id == "." || id == ".." {
		return fmt.Errorf("invalid id")
	}
	if err := os.WriteFile(filepath.Join(s.Dir, id), []byte("queued\n"), 0600); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.NotifyURL, strings.NewReader(id))
	if err != nil {
		return err
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("notification status %d", resp.StatusCode)
	}
	return nil
}
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := s.Submit(r.Context(), r.URL.Query().Get("id")); err != nil {
		http.Error(w, "submission failed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
