// Package recordfmt formats records for text reports.
package recordfmt

// Formatter formats key/value records with a prefix.
type Formatter struct{ prefix string }

// New creates a formatter with the supplied prefix.
func New(prefix string) *Formatter { return &Formatter{prefix: prefix} }

// Format returns one record without a trailing newline.
func (f *Formatter) Format(key, value string) string { return f.prefix + key + ":" + value }
