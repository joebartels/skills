package recordload

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// Record is a key=value pair.
type Record struct{ Key, Value string }

// LineError reports a malformed record. Line is the one-based physical line
// number, and Text is the unmodified line without its newline separator.
type LineError struct {
	Line int
	Text string
}

func (e *LineError) Error() string {
	return fmt.Sprintf("line %d: malformed record", e.Line)
}

// Load reads key=value records in order, ignoring empty lines and lines
// beginning with #. It returns the valid prefix on a malformed record or reader
// failure. Malformed records are reported as *LineError, joined with any reader
// failure. Load does not close r.
func Load(r io.Reader) ([]Record, error) {
	data, readErr := io.ReadAll(r)
	var records []Record
	for n, text := range strings.Split(string(data), "\n") {
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		key, value, ok := strings.Cut(text, "=")
		if !ok {
			lineErr := &LineError{Line: n + 1, Text: text}
			if readErr != nil {
				return records, errors.Join(lineErr, readErr)
			}
			return records, lineErr
		}
		records = append(records, Record{Key: key, Value: value})
	}
	return records, readErr
}
