// Package roster summarizes tenant membership.
package roster

// Entry associates a display name with a tenant.
type Entry struct {
	Tenant string
	Name   string
}

// Summary describes records and their distinct display names.
type Summary struct {
	Total int
	Names []string
}
