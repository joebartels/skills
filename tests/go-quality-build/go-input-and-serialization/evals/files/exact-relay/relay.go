package relay

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// MaxDocumentBytes is the maximum input size, including whitespace.
const MaxDocumentBytes = 4096

// Relay normalizes an event's kind and publishes its routing id and JSON body.
func Relay(r io.Reader, publish func(uint64, []byte) error) error {
	var event map[string]any
	if err := json.NewDecoder(io.LimitReader(r, MaxDocumentBytes)).Decode(&event); err != nil {
		return fmt.Errorf("decode event: %w", err)
	}
	id, ok := event["id"].(float64)
	if !ok {
		return fmt.Errorf("event id must be a number")
	}
	kind, ok := event["kind"].(string)
	if !ok {
		return fmt.Errorf("event kind must be a string")
	}
	body := map[string]any{"id": id, "kind": strings.ToLower(strings.TrimSpace(kind))}
	if extensions, present := event["extensions"]; present {
		body["extensions"] = extensions
	}
	wire, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode event: %w", err)
	}
	if err := publish(uint64(id), wire); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}
