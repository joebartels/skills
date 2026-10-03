// Package ledger loads textual ledger records into caller-managed maps.
package ledger

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrInvalidRow indicates a row with an invalid format or value.
// Callers can detect it with errors.Is.
var ErrInvalidRow = errors.New("invalid row")

// Load reads records from r into destination and returns the accepted count.
// Each line must be name=value, with a nonempty name and a nonnegative integer
// value. The destination map must be non-nil.
//
// Repeated names replace earlier values and count as additional accepted
// records. On failure, Load returns the accepted count and leaves earlier
// writes in destination.
//
// Load does not close r. The scanner's default token limit bounds line size.
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
