package keycodec

import (
	"errors"
	"net/url"
	"strings"
)

// ErrInvalidKey indicates an empty field or an invalid wire representation.
var ErrInvalidKey = errors.New("invalid key")

// Key identifies a name within a region. Both fields must be non-empty.
type Key struct {
	Region string
	Name   string
}

// Encode returns two individually URL path-escaped fields separated by a slash.
// Empty fields return an empty string and ErrInvalidKey.
func Encode(k Key) (string, error) {
	if k.Region == "" || k.Name == "" {
		return "", ErrInvalidKey
	}
	return url.PathEscape(k.Region) + "/" + url.PathEscape(k.Name), nil
}

// Decode unescapes the two fields of a wire key. A malformed escape, empty
// field, or incorrect segment count returns a zero Key and ErrInvalidKey.
func Decode(wire string) (Key, error) {
	regionWire, nameWire, found := strings.Cut(wire, "/")
	if !found || regionWire == "" || nameWire == "" || strings.Contains(nameWire, "/") {
		return Key{}, ErrInvalidKey
	}
	region, err := url.PathUnescape(regionWire)
	if err != nil {
		return Key{}, ErrInvalidKey
	}
	name, err := url.PathUnescape(nameWire)
	if err != nil {
		return Key{}, ErrInvalidKey
	}
	return Key{Region: region, Name: name}, nil
}
