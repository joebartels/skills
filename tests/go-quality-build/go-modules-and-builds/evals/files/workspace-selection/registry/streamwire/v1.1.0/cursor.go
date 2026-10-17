package streamwire

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
)

// ErrCursor identifies a malformed resume cursor.
var ErrCursor = errors.New("invalid cursor")

// EncodeCursor encodes the service's versioned, URL-safe resume cursor.
func EncodeCursor(topic string, offset uint64) (string, error) {
	if err := ValidateTopic(topic); err != nil {
		return "", err
	}
	return "sc1." + base64.RawURLEncoding.EncodeToString([]byte(topic)) + "." + strconv.FormatUint(offset, 36), nil
}

// DecodeCursor recovers the topic and offset from a canonical resume cursor.
func DecodeCursor(token string) (string, uint64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "sc1" {
		return "", 0, ErrCursor
	}
	topic, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || ValidateTopic(string(topic)) != nil {
		return "", 0, ErrCursor
	}
	offset, err := strconv.ParseUint(parts[2], 36, 64)
	if err != nil {
		return "", 0, ErrCursor
	}
	canonical, _ := EncodeCursor(string(topic), offset)
	if token != canonical {
		return "", 0, ErrCursor
	}
	return string(topic), offset, nil
}
