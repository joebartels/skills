// Package measure is internal to a CLI with no external consumers.
package measure

import "errors"

var errInvalid = errors.New("invalid measurement")

type Measurement struct {
	Value  int
	Legacy bool
}

// WithValue returns a modified copy.
func (m Measurement) WithValue(value int) Measurement {
	m.Value = value
	return m
}

// Validate checks the current format; ValidateLegacy also permits old records.
func Validate(m Measurement) error { return validate(m, false) }

func ValidateLegacy(m Measurement) error { return validate(m, true) }

func validate(m Measurement, allowLegacy bool) error {
	if m.Value < 0 || (m.Legacy && !allowLegacy) {
		return errInvalid
	}
	return nil
}

// Total accepts both input formats and returns a current-format measurement.
// It is the CLI's only production entry point.
func Total(records []Measurement) (Measurement, error) {
	total := 0
	for _, m := range records {
		if err := ValidateLegacy(m); err != nil {
			return Measurement{}, err
		}
		total += m.Value
	}
	return Measurement{Value: total}, nil
}
