package fielddesk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Bulletin struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Service struct {
	root string
}

func New(root string) *Service {
	return &Service{root: root}
}

func (s *Service) SubmitCheckin(ctx context.Context, id, note string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !validID(id) || strings.TrimSpace(note) == "" {
		return errors.New("invalid check-in")
	}
	directory := filepath.Join(s.root, "checkins")
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, id+".txt"), []byte(note+"\n"), 0644)
}

func validID(id string) bool {
	if id == "" {
		return false
	}
	for _, r := range id {
		if r != '-' && (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}

func (s *Service) Bulletins(ctx context.Context) ([]Bulletin, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(s.root, "bulletins.json"))
	if errors.Is(err, os.ErrNotExist) {
		return []Bulletin{}, nil
	}
	if err != nil {
		return nil, err
	}
	var bulletins []Bulletin
	if err := json.Unmarshal(data, &bulletins); err != nil {
		return nil, fmt.Errorf("decode bulletins: %w", err)
	}
	if bulletins == nil {
		return []Bulletin{}, nil
	}
	return bulletins, nil
}

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/checkins", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var input struct {
			ID   string `json:"id"`
			Note string `json:"note"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid check-in", http.StatusBadRequest)
			return
		}
		if err := s.SubmitCheckin(r.Context(), input.ID, input.Note); err != nil {
			if !validID(input.ID) || strings.TrimSpace(input.Note) == "" {
				http.Error(w, "invalid check-in", http.StatusBadRequest)
				return
			}
			http.Error(w, "cannot save check-in", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(struct {
			ID string `json:"id"`
		}{ID: input.ID})
	})
	mux.HandleFunc("/v1/bulletins", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		bulletins, err := s.Bulletins(r.Context())
		if err != nil {
			http.Error(w, "cannot read bulletins", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(bulletins)
	})
	return mux
}
