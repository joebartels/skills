package catalog

import (
	"encoding/json"
	"net/http"
)

type Item struct {
	SKU   string `json:"sku"`
	Price int    `json:"price"`
}
type Service struct{ Items map[string]Item }

func (s *Service) Lookup(sku string) (Item, bool) { v, ok := s.Items[sku]; return v, ok }
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	item, ok := s.Lookup(r.URL.Query().Get("sku"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(item); err != nil {
		return
	}
}
