# cacheworker

This is the internal cache expiry helper from a single private worker binary.
The helper is unexported, has no persisted or serialized representation, and
is called only within its package. Go 1.22 is supported.
