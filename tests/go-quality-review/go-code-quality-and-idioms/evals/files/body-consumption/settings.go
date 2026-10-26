package settings

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var (
	errTooLarge = errors.New("settings body too large")
	errInvalid  = errors.New("invalid settings body")
)

// Handler accepts one settings object of at most 1024 bytes, including trailing
// whitespace. It returns 413 for any over-limit body and applies only valid input.
// The owning router authorizes callers before invoking this handler.
type Handler struct {
	Apply func(map[string]string)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	settings, err := decode(http.MaxBytesReader(w, r.Body, 1024))
	if errors.Is(err, errTooLarge) {
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.Apply(settings)
	w.WriteHeader(http.StatusNoContent)
}

func decode(r io.Reader) (map[string]string, error) {
	d := json.NewDecoder(r)
	var settings map[string]string
	err := d.Decode(&settings)
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return nil, errTooLarge
	}
	if err != nil || settings == nil {
		return nil, errInvalid
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, errInvalid
	}
	return settings, nil
}
