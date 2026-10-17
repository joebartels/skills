# Severity registry

This Go 1.22.0 library publishes Name(code int) string. Extend its protocol registry with warn=30 and error=40, preserving debug=10, info=20, and the empty string for unknown codes. schema.txt is the source of truth. Both the default build and the compact build tag are supported on linux/amd64 and linux/arm64.

Keep generated tables checked in so ordinary users can build without running a generator. Maintainers must also be able to regenerate the tables from a fresh checkout with just the supported Go toolchain; repeated generation must leave identical Go output. The current generator predates the compact variant. Extend the existing generator and add focused tests; retain both implementations and the existing public API.
