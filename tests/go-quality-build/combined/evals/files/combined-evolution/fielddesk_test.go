package fielddesk_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/fielddesk"
)

func TestCheckinCallerAndHTTPContract(t *testing.T) {
	root := t.TempDir()
	s := fielddesk.New(root)
	if err := s.SubmitCheckin(context.Background(), "crew-1", "arrived"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "checkins", "crew-1.txt"))
	if err != nil || string(data) != "arrived\n" {
		t.Fatalf("stored check-in = %q, %v", data, err)
	}
	request := httptest.NewRequest(http.MethodPost, "/v1/checkins", strings.NewReader(`{"id":"crew-2","note":"ready"}`))
	response := httptest.NewRecorder()
	s.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusCreated || response.Body.String() != "{\"id\":\"crew-2\"}\n" {
		t.Fatalf("response = %d %q", response.Code, response.Body.String())
	}
	request = httptest.NewRequest(http.MethodPost, "/v1/checkins", strings.NewReader(`{"id":"../bad","note":"bad"}`))
	response = httptest.NewRecorder()
	s.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("invalid check-in status = %d", response.Code)
	}
}

func TestBulletinCallerAndHTTPContract(t *testing.T) {
	root := t.TempDir()
	s := fielddesk.New(root)
	bulletins, err := s.Bulletins(context.Background())
	if err != nil || bulletins == nil || len(bulletins) != 0 {
		t.Fatalf("absent snapshot = %#v, %v", bulletins, err)
	}
	request := httptest.NewRequest(http.MethodGet, "/v1/bulletins", nil)
	response := httptest.NewRecorder()
	s.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "[]\n" || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("empty response = %d %q %q", response.Code, response.Body.String(), response.Header().Get("Content-Type"))
	}
	if err := os.WriteFile(filepath.Join(root, "bulletins.json"), []byte(`[{"id":"b-1","text":"north gate"}]`), 0644); err != nil {
		t.Fatal(err)
	}
	bulletins, err = s.Bulletins(context.Background())
	if err != nil || len(bulletins) != 1 || bulletins[0].Text != "north gate" {
		t.Fatalf("stored snapshot = %#v, %v", bulletins, err)
	}
}
