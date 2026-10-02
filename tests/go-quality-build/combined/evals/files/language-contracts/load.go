package batchview

import (
	"errors"
	"io"
	"strings"
)

type Item struct {
	Key  string            `json:"key"`
	Tags map[string][]byte `json:"tags"`
}

func Load(r io.Reader) ([]Item, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var out []Item
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return out, errors.New("malformed item")
		}
		out = append(out, Item{Key: key, Tags: map[string][]byte{"data": []byte(value)}})
	}
	return out, nil
}
