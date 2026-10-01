package reportkit

import (
	"bytes"
	"encoding/json"
	"io"
)

type Record struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Encoder is the report-format extension protocol implemented by hosts and bundled formats.
type Encoder interface {
	Encode(io.Writer, []Record) error
}

type JSON struct{}

func (JSON) Encode(w io.Writer, records []Record) error { return json.NewEncoder(w).Encode(records) }

type Report struct{ Encoder Encoder }

func (r Report) Render(records []Record) ([]byte, error) {
	var b bytes.Buffer
	if err := r.Encoder.Encode(&b, records); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
