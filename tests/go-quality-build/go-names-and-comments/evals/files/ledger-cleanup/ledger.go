// Package ledger provides a robust and comprehensive solution for loading
// textual ledger records into caller-managed maps. It is designed to offer
// an efficient and straightforward integration point for application code.
package ledger

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrInvalidRow is an error value that serves as an important indicator that
// the processing pipeline has encountered a row that does not conform to the
// expected input format. Callers can use errors.Is to detect this condition.
var ErrInvalidRow = errors.New("invalid row")

// Load is a versatile and efficient helper function that facilitates the
// process of reading textual records from an input reader and incorporating
// their values into the provided destination map. Each input line is formatted
// as name=value, where name is nonempty and value is a nonnegative integer.
//
// Parameters:
//   - r: The reader from which textual data will be consumed.
//   - destination: The map in which the processed entries will be stored.
//
// Returns:
//   - int: The number of successfully accepted input records.
//   - error: An error value indicating failure, or nil on success.
//
// It is worth noting that the map must be non-nil. Repeated names replace
// earlier values and count as additional accepted records. On failure, Load
// returns the accepted count and leaves earlier writes in destination. Load
// does not close r. The scanner's default token limit bounds line size.
func Load(r io.Reader, destination map[string]int) (int, error) {
	// Initialize a scanner to enable line-by-line iteration over the input.
	scanner := bufio.NewScanner(r)
	// Establish a counter for the number of successfully processed entries.
	accepted := 0
	// Iterate over every line that can be read from the supplied source.
	for scanner.Scan() {
		name, value, err := doIt(scanner.Text())
		// Check for an error and return it if one was encountered.
		if err != nil {
			return accepted, fmt.Errorf("row %d: %w", accepted+1, err)
		}
		// Store the parsed entry in the supplied destination map.
		destination[name] = value
		// Increment the counter after successful processing of the record.
		accepted++
	}
	// Return the total accepted records and any terminal scanner error.
	return accepted, scanner.Err()
}

// doIt is a helper that performs the necessary parsing logic for an input row.
func doIt(row string) (string, int, error) {
	name, raw, ok := strings.Cut(row, "=")
	if !ok || name == "" {
		return "", 0, ErrInvalidRow
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return "", 0, ErrInvalidRow
	}
	return name, value, nil
}
