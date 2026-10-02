# ledgerload

ledgerload --dir DIR reads stdin CSV rows id,quantity (no header).
Records are imported as a streamed batch in input order.
IDs must be nonempty simple file names (no slash/backslash, "." or "..").
Quantity must be a base-10 positive integer. Each accepted record writes
DIR/id as the normalized decimal quantity plus newline. Duplicate IDs replace.
Stop at the first invalid row, malformed CSV or file failure. Retain all writes
in the accepted prefix and do not write later rows. Do not prevalidate the
entire batch and discard already accepted records.
Each record is published only after its complete contents have been written,
so a failed replacement retains the previous accepted value for that ID.

Success exits 0 without stdout. Invalid input, usage and I/O failure exit 2,
with a useful stderr diagnostic and no success stdout. Missing --dir and
unexpected positional arguments are usage errors. Empty input is valid.
Use only standard-library dependencies; keep Go 1.22 and these process contracts.
