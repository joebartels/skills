package recordload

import (
	"fmt"
	"io"
	"strings"
)

type Record struct{ Key, Value string }

func Load(r io.Reader) ([]Record, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var records []Record
	for n, text := range strings.Split(string(data), "\n") {
		if text == "" {
			continue
		}
		key, value, ok := strings.Cut(text, "=")
		if !ok {
			return records, fmt.Errorf("line %d: malformed record", n+1)
		}
		records = append(records, Record{Key: key, Value: value})
	}
	return records, nil
}
