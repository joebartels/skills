package keycodec

import (
	"errors"
	"net/url"
	"strings"
)

// ErrInvalidKey indicates an empty field or invalid encoded key.
var ErrInvalidKey = errors.New("invalid key")

// Key identifies a named resource within a region. Both fields must be nonempty.
type Key struct {
	Region string
	Name   string
}

// Encode escapes each field as a URL path segment and joins them with a slash.
// It returns an empty string and ErrInvalidKey if either field is empty.
func Encode(k Key) (string, error) {
	if k.Region == "" || k.Name == "" {
		return "", ErrInvalidKey
	}
	return url.PathEscape(k.Region) + "/" + url.PathEscape(k.Name), nil
}

// Decode unescapes two nonempty URL path segments separated by one slash.
// It returns a zero Key and ErrInvalidKey for an invalid segment count,
// an empty field, or a malformed percent escape. Each field is unescaped once.
func Decode(wire string) (Key, error) {
	region, name, ok := strings.Cut(wire, "/")
	if !ok || region == "" || name == "" || strings.Contains(name, "/") {
		return Key{}, errors.New("invalid key")
	}
	region, err := url.PathUnescape(region)
	if err != nil {
		return Key{}, errors.New("invalid key")
	}
	name, err = url.PathUnescape(name)
	if err != nil {
		return Key{}, errors.New("invalid key")
	}
	return Key{Region: region, Name: name}, nil
}
