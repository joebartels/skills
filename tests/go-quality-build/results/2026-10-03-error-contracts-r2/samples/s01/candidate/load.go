package recordload

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Record is a key and value read from one line.
type Record struct{ Key, Value string }

// LineError describes a malformed record.
type LineError struct {
	Line int    // One-based physical line number, including blanks and comments.
	Text string // Original line text, without its newline.
}

func (e *LineError) Error() string {
	return fmt.Sprintf("line %d: malformed record", e.Line)
}

// Load reads key=value records in order from a borrowed reader.
// Empty lines and lines beginning with # are ignored. Fields are not trimmed.
// On failure, Load returns the valid prefix and an error exposing any LineError
// and reader cause through errors.As and errors.Is.
func Load(r io.Reader) ([]Record, error) {
	data, err := io.ReadAll(r)
	var records []Record
	for n, text := range strings.Split(string(data), "\n") {
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		key, value, ok := strings.Cut(text, "=")
		if !ok {
			lineErr := &LineError{Line: n + 1, Text: text}
			if err != nil {
				return records, errors.Join(lineErr, err)
			}
			return records, lineErr
		}
		records = append(records, Record{Key: key, Value: value})
	}
	return records, err
}
