package keycodec

import (
	"errors"
	"net/url"
	"strings"
)

// ErrInvalidKey indicates an empty field, malformed escape, or invalid segment count.
var ErrInvalidKey = errors.New("invalid key")

// Key identifies a name within a region. Both fields must be nonempty.
type Key struct {
	Region string
	Name   string
}

// Encode joins the individually URL path-escaped fields with one slash.
// On failure it returns an empty string and ErrInvalidKey.
func Encode(k Key) (string, error) {
	if k.Region == "" || k.Name == "" {
		return "", ErrInvalidKey
	}
	return url.PathEscape(k.Region) + "/" + url.PathEscape(k.Name), nil
}

// Decode splits a wire key into two segments, then URL path-unescapes each once.
// On failure it returns a zero Key and ErrInvalidKey.
func Decode(wire string) (Key, error) {
	p := strings.Split(wire, "/")
	if len(p) != 2 || p[0] == "" || p[1] == "" {
		return Key{}, ErrInvalidKey
	}
	region, err := url.PathUnescape(p[0])
	if err != nil {
		return Key{}, ErrInvalidKey
	}
	name, err := url.PathUnescape(p[1])
	if err != nil {
		return Key{}, ErrInvalidKey
	}
	return Key{Region: region, Name: name}, nil
}
