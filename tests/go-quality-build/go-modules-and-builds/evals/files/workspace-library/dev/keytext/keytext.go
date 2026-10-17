// Package keytext defines catalog name formatting policy.
package keytext

import (
	"errors"
	"strings"
)

// ErrName identifies an invalid catalog name.
var ErrName = errors.New("invalid catalog name")

// Display removes surrounding whitespace from a displayed name.
func Display(name string) string {
	return strings.TrimSpace(name)
}

// Canonical lowercases ASCII words and joins them with hyphens.
func Canonical(name string) (string, error) {
	words := strings.Fields(name)
	if len(words) == 0 {
		return "", ErrName
	}
	for _, word := range words {
		for _, r := range word {
			if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
				continue
			}
			return "", ErrName
		}
	}
	return strings.ToLower(strings.Join(words, "-")), nil
}
