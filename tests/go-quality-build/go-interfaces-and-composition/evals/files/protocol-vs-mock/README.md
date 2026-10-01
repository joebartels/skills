# reportkit

This package supplies report formats and the extension protocol used by host applications. Hosts implement Encoder for proprietary formats; the shared Record and Encoder names are part of the supported library API. Encode writes records in order and reports write/encoding errors. Implementations must accept empty input. Callers may use Report literals with any Encoder. Report renders into a private buffer and returns no bytes when encoding fails.

JSON is newline-terminated and currently the only bundled format. The requested CSV format uses the columns name,value, includes a header even for no records, and follows encoding/csv quoting rules. A value is text, not a number. There is no separate Report implementation in production. The API is small enough to document in this file; Go 1.22 is supported.
