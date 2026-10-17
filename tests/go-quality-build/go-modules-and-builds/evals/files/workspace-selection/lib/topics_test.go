package streamkit_test

import (
	"errors"
	"testing"

	"example.org/streamkit"
	"example.org/streamwire"
)

func TestTopicPath(t *testing.T) {
	got, err := streamkit.TopicPath("orders.eu-west")
	if err != nil || got != "/topics/orders.eu-west" {
		t.Fatalf("TopicPath = %q, %v", got, err)
	}
	for _, topic := range []string{"", "Orders", " orders", "a..b"} {
		got, err := streamkit.TopicPath(topic)
		if got != "" || !errors.Is(err, streamwire.ErrTopic) || err == streamwire.ErrTopic {
			t.Fatalf("TopicPath(%q) = %q, %v; want empty path and wrapped ErrTopic", topic, got, err)
		}
	}
}
