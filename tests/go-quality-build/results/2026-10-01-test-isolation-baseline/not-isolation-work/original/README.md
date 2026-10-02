# Pure clamp repair

`Clamp(value, low, high int) int` returns low for values below the inclusive
range, high for values above it, or value when it is in the inclusive range.
Callers supply low <= high. The lower-bound branch is wrong. Repair it and
add a focused ordinary regression test, retaining the public API and Go 1.22.
This pure calculation has no I/O, process state, dependencies or concurrency.
