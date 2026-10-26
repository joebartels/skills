package settings

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		code int
	}{
		{"valid", `{"theme":"dark"}`, http.StatusNoContent},
		{"invalid", `!`, http.StatusBadRequest},
		{"large value", `{"theme":"` + strings.Repeat("x", 2048) + `"}`, http.StatusRequestEntityTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			applied := false
			h := Handler{Apply: func(map[string]string) { applied = true }}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(tc.body)))
			if w.Code != tc.code || applied != (tc.code == http.StatusNoContent) {
				t.Fatalf("status=%d applied=%t", w.Code, applied)
			}
		})
	}
}
