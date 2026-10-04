package accountlookup

import (
	"errors"
	"fmt"
	"strings"
)

var ErrMissing = errors.New("account missing")
var ErrUnavailable = errors.New("account unavailable")

// ErrInvalidKey indicates an empty or whitespace-only account key.
var ErrInvalidKey = errors.New("invalid account key")

var BackendMissing = errors.New("backend missing")

type Backend interface{ Find(string) (string, error) }
type Service struct{ Backend Backend }

// Lookup returns an account value, or an empty value on failure.
// Missing accounts return ErrMissing directly; unavailable errors wrap
// ErrUnavailable with lookup and quoted-key context, without backend diagnostics.
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
