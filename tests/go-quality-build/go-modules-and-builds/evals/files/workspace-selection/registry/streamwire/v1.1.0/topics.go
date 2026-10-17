// Package streamwire implements the stream service wire values.
package streamwire

import (
	"errors"
	"strings"
)

// ErrTopic identifies invalid topic names.
var ErrTopic = errors.New("invalid topic")

// ValidateTopic checks the service's case-sensitive topic grammar.
func ValidateTopic(topic string) error {
	if len(topic) == 0 || len(topic) > 128 {
		return ErrTopic
	}
	for _, label := range strings.Split(topic, ".") {
		if len(label) == 0 || label[0] < 'a' || label[0] > 'z' {
			return ErrTopic
		}
		for i := 1; i < len(label); i++ {
			c := label[i]
			if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
				continue
			}
			return ErrTopic
		}
	}
	return nil
}
