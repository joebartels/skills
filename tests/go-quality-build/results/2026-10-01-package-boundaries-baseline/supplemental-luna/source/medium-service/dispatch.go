package dispatch

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Service processes one delivery request at a time. Callers serialize access.
type Service struct {
	Dir, NotifyURL string
	Client         *http.Client
	Store          RecordStore
	Notifier       Notifier
}

// RecordStore persists the bytes for an accepted record.
type RecordStore interface {
	Save(context.Context, string, []byte) error
}

// Notifier sends an accepted ID to the notification service.
type Notifier interface {
	Notify(context.Context, string) error
}

type fileStore struct{ dir string }

func (s fileStore) Save(_ context.Context, id string, record []byte) error {
	return os.WriteFile(filepath.Join(s.dir, id), record, 0600)
}

type httpNotifier struct {
	url    string
	client *http.Client
}

func (n httpNotifier) Notify(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.url, strings.NewReader(id))
	if err != nil {
		return err
	}
	resp, err := n.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("notification status %d", resp.StatusCode)
	}
	return nil
}

// Submit validates an ID, writes its record, then notifies the external service.
// A notification failure leaves the record in place and returns an error; no retry is promised.
func (s *Service) Submit(ctx context.Context, id string) error {
	if id == "" || strings.ContainsAny(id, "/\\") || id == "." || id == ".." {
		return fmt.Errorf("invalid id")
	}
	store := s.Store
	if store == nil {
		store = fileStore{dir: s.Dir}
	}
	if err := store.Save(ctx, id, []byte("queued\n")); err != nil {
		return err
	}
	notifier := s.Notifier
	if notifier == nil {
		notifier = httpNotifier{url: s.NotifyURL, client: s.Client}
	}
	return notifier.Notify(ctx, id)
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

// Replay submits each nonblank trimmed line in order and stops at the first
// read or submission error.
func (s *Service) Replay(ctx context.Context, input io.Reader) error {
	scanner := bufio.NewScanner(input)
	line := 0
	for scanner.Scan() {
		line++
		id := strings.TrimSpace(scanner.Text())
		if id == "" {
			continue
		}
		if err := s.Submit(ctx, id); err != nil {
			return fmt.Errorf("line %d: %w", line, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read replay file: %w", err)
	}
	return nil
}
