# Catalog change for skill evaluation

This module is a public v1.4 library. `Catalog.IDs` promises a stable-order,
caller-owned snapshot: callers may sort or redact returned IDs without changing
the Catalog. The Catalog can be shared among goroutines.

The proposed change replaces the previous `slices.Clone(c.ids)` return with a
full slice expression intended to prevent callers from changing the Catalog.
No other implementation changed.
