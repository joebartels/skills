# Partial configuration updates

This Go 1.22 package applies configuration changes using standard `encoding/json`.
Extend `Apply` while preserving its API. Accept exactly one JSON object whose
case-sensitive optional members are `enabled`, `retries`, `label`, and `labels`.
Reject unknown members and duplicate member names after JSON escape decoding,
including duplicate keys inside `labels`.

Absent members retain their existing values. Explicit `false`, `0`, and empty
strings are updates. `enabled` must be a boolean; `retries` must be an integer
from 0 through 10; neither permits null. `label` accepts a string or null (clear).
`labels` accepts an object of string values or null (clear). A supplied labels
object replaces the whole map; `{}` leaves an empty non-nil map. Map values
cannot be null. Surrounding whitespace is allowed; any other extra input fails.

Apply the update atomically. On any input or validation error, preserve every
original field and all maps or pointers held by other callers. On success,
updating a label or replacing labels must also leave those old aliases intact.
