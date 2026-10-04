package accountlookup

import (
	"errors"
	"fmt"
	"strings"
)

var ErrMissing = errors.New("account missing")
var ErrUnavailable = errors.New("account unavailable")
var BackendMissing = errors.New("backend missing")

// ErrInvalidKey indicates an empty or whitespace-only account key.
var ErrInvalidKey = errors.New("invalid account key")

type Backend interface{ Find(string) (string, error) }
type Service struct{ Backend Backend }

// Lookup returns an account for key, passing nonblank keys to Backend unchanged.
// Blank keys return ErrInvalidKey without calling Backend. Missing accounts return
// ErrMissing directly; other backend failures wrap ErrUnavailable with safe context.
// The returned value is usable only when the error is nil.
func (s Service) Lookup(key string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", ErrInvalidKey
	}
	value, err := s.Backend.Find(key)
	if errors.Is(err, BackendMissing) {
		return "", ErrMissing
	}
	if err != nil {
		return "", fmt.Errorf("lookup account %q: %w", key, ErrUnavailable)
	}
	return value, nil
}
