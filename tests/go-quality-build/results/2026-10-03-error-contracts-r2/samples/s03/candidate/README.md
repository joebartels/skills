# accountlookup

A Go 1.22 library for looking up account values through a supplied `Backend`.
`Service.Lookup` rejects empty or Unicode-whitespace-only keys with
`ErrInvalidKey` before calling the backend. Nonblank keys reach the backend
unchanged. Use the returned value only when the error is nil; every failure
returns an empty value.

Absent accounts return `ErrMissing` directly, preserving `err == ErrMissing`.
Other backend failures wrap `ErrUnavailable` with the lookup operation and quoted
key. Backend error identity, type, and diagnostic text remain private. Quoting
escapes control characters in the key.

```go
value, err := service.Lookup(key)
switch {
case err == accountlookup.ErrMissing:
	// Account does not exist.
case errors.Is(err, accountlookup.ErrInvalidKey):
	// Supply a nonblank key.
case errors.Is(err, accountlookup.ErrUnavailable):
	// Retry or report the safe lookup context in err.
case err == nil:
	// Use value.
}
```
