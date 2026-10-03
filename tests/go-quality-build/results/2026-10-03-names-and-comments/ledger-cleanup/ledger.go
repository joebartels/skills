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

// ErrInvalidRow indicates an invalid input row. Use errors.Is to detect it.
var ErrInvalidRow = errors.New("invalid row")

// Load reads name=value records from r into destination and returns the number
// of accepted records. Names must be nonempty and values must be nonnegative
// integers.
//
// destination must be non-nil. Repeated names replace earlier values and count
// as additional accepted records.
//
// On failure, Load returns the accepted count and leaves earlier writes in
// destination. It does not close r. Line size is bounded by bufio.Scanner's
// default token limit.
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
