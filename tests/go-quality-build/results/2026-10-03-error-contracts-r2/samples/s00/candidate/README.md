# accountlookup

Go 1.22 account lookup library. `Service.Lookup` rejects empty and Unicode
whitespace-only keys with `ErrInvalidKey` before calling `Backend`. Nonblank keys
are passed through unchanged.

Missing accounts return `ErrMissing` directly, preserving `err == ErrMissing`.
Other backend failures wrap `ErrUnavailable` with the lookup operation and quoted
key. Backend error identity, types, and diagnostics remain private. Key control
characters are escaped in the context. Use the returned value only if `err == nil`.

```go
value, err := service.Lookup(key)
switch {
case err == nil:
    // Use value.
case err == accountlookup.ErrMissing:
    // Account does not exist.
case errors.Is(err, accountlookup.ErrInvalidKey):
    // Supply a nonblank key.
case errors.Is(err, accountlookup.ErrUnavailable):
    // Backend lookup failed; err contains safe operation and key context.
}
```
