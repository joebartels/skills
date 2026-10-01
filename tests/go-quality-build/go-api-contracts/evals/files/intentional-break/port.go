// Package portnum parses numeric port settings.
package portnum

import "strconv"

// Parse returns the port, or zero when text is invalid.
func Parse(text string) int {
	n, err := strconv.Atoi(text)
	if err != nil || n < 0 || n > 65535 {
		return 0
	}
	return n
}
