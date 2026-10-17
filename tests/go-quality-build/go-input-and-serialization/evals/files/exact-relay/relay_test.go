package relay_test

import (
	"encoding/json"
	"strings"
	"testing"

	relay "example.com/exactrelay"
)

func TestRelaySmallEvent(t *testing.T) {
	calls := 0
	err := relay.Relay(strings.NewReader(`{"id":17,"kind":" Created ","extensions":{"source":"web"}}`), func(id uint64, body []byte) error {
		calls++
		var got struct {
			ID         uint64            `json:"id"`
			Kind       string            `json:"kind"`
			Extensions map[string]string `json:"extensions"`
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatal(err)
		}
		if id != 17 || got.ID != 17 || got.Kind != "created" || got.Extensions["source"] != "web" {
			t.Fatalf("routing id %d, event %+v", id, got)
		}
		return nil
	})
	if err != nil || calls != 1 {
		t.Fatalf("Relay error %v, publish calls %d", err, calls)
	}
}
