// Package codes validates the identifiers accepted by billing import tools.
package codes

//go:generate go run ./internal/cataloggen

// Valid reports whether value is an identifier of the named format.
// Format names are case-sensitive; unknown formats reject every value.
func Valid(name, value string) bool {
	rule, ok := lookup(name)
	if !ok || len(value) != len(rule.prefix)+rule.digits {
		return false
	}
	if value[:len(rule.prefix)] != rule.prefix {
		return false
	}
	for _, c := range value[len(rule.prefix):] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

type rule struct {
	prefix string
	digits int
}
