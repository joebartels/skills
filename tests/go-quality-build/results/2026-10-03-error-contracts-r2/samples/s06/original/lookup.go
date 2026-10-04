package accountlookup

import "errors"

var ErrMissing = errors.New("account missing")
var ErrUnavailable = errors.New("account unavailable")
var BackendMissing = errors.New("backend missing")

type Backend interface{ Find(string) (string, error) }
type Service struct{ Backend Backend }

func (s Service) Lookup(key string) (string, error) {
	value, err := s.Backend.Find(key)
	if errors.Is(err, BackendMissing) {
		return "", ErrMissing
	}
	if err != nil {
		return "", ErrUnavailable
	}
	return value, nil
}
