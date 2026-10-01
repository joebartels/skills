package consumer_test

import (
	"example.com/recordfmt"
	"testing"
)

var factories = map[string]func(string) *recordfmt.Formatter{
	"text": recordfmt.New,
}

func TestRegisteredFactory(t *testing.T) {
	if got := factories["text"]("p/").Format("key", "value"); got != "p/key:value" {
		t.Fatalf("record = %q", got)
	}
}
