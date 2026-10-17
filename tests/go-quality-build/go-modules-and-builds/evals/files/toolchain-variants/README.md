# Roster reports

Add per-tenant record counts to the roster summary as the exported `TenantCounts map[string]int` field. Every record contributes to its tenant's count, including a record whose Name is empty. Keep the existing total and sorted distinct nonempty names. An empty report has zero total, a non-nil empty Names slice, and a non-nil empty TenantCounts map. Do not mutate input records.

The package is imported by the scheduler and the audit command. See docs/development.md for development commands.
