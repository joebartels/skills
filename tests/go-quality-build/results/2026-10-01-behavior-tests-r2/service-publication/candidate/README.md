# Local item mirror

Add remote refresh to this simple snapshot library. Refresh(ctx, client,
endpoint, path) makes a GET through the supplied client/context, accepts a 200
JSON array of Item{Code, Label}, validates each code and label as nonblank
after trimming, then publishes a JSON array with both lowercase fields.
Do not discard or normalize nonblank field values. Non-200 status, transport
failure, malformed JSON, trailing JSON and invalid items must return an error
and leave exact previous file bytes intact. Empty arrays are valid.

Readers must see complete snapshots: publish by replacement, without
truncating an existing inode. On supported Unix filesystems an already-open
old handle retains the old snapshot, while a new open sees the full new one.
Do not claim equivalent semantics on Windows without checking them.
Close response bodies. Keep Refresh's existing signature, Go 1.22 and standard
library dependencies; no prescribed storage interface or testing framework.
