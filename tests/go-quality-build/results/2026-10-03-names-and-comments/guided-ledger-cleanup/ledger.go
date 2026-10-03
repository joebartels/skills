// Package ledger loads textual records into caller-managed maps.
package ledger

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrInvalidRow indicates a malformed row and can be detected with errors.Is.
var ErrInvalidRow = errors.New("invalid row")

// Load reads name=value lines into destination and returns the accepted count.
// Names must be nonempty, values must be nonnegative integers, and destination
// must be non-nil. Repeated names replace earlier values and count as additional
// accepted records.
//
// On failure, Load returns the accepted count and leaves earlier writes in
// destination. Invalid rows produce errors wrapping ErrInvalidRow. Load
// does not close r. The scanner's default token limit bounds line size.
func Load(r io.Reader, destination map[string]int) (int, error) {
	scanner := bufio.NewScanner(r)
	accepted := 0
	for scanner.Scan() {
		name, value, err := parseRow(scanner.Text())
		if err != nil {
			return accepted, fmt.Errorf("row %d: %w", accepted+1, err)
		}
		destination[name] = value
		accepted++
	}
	return accepted, scanner.Err()
}

func parseRow(row string) (string, int, error) {
	name, rawValue, ok := strings.Cut(row, "=")
	if !ok || name == "" {
		return "", 0, ErrInvalidRow
	}
	value, err := strconv.Atoi(rawValue)
	if err != nil || value < 0 {
		return "", 0, ErrInvalidRow
	}
	return name, value, nil
}
