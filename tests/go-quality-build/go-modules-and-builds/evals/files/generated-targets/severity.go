// Package severity names wire protocol levels.
package severity

//go:generate go run ./internal/gen

// Name returns the protocol name, or an empty string for an unknown code.
func Name(code int) string { return lookup(code) }
