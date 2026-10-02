# File-backed store test work

`New(root string) *Store`, `(*Store).Put(key, value string) error` and
`(*Store).Get(key string) (string, error)` use only the supplied root. Supported
keys contain ASCII letters, digits, underscore or hyphen and are nonempty.
Put replaces the key's complete string; Get returns it unchanged or an error
for an invalid/missing key. Different roots are independent even for identical
keys. Concurrent operations on different roots are supported; same-key
concurrent replacement within one root is outside this task.

The implementation already supports these contracts. Extend the tests to
cover concurrent independent instances and grouped child cases, including
repeated same-key values in different roots. Test-owned files/resources must
remain usable until their work finishes. Keep useful serial checks. Public
API, standard-library dependencies and Go 1.22 minimum must remain unchanged.
