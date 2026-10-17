// Package streamkit provides request values for stream service clients.
package streamkit

import (
	"fmt"

	"example.org/streamwire"
)

// TopicPath returns the service path for a topic.
func TopicPath(topic string) (string, error) {
	if err := streamwire.ValidateTopic(topic); err != nil {
		return "", fmt.Errorf("topic path: %w", err)
	}
	return "/topics/" + topic, nil
}
