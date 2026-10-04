# recordload
Go 1.22 library. Load reads key=value records in order. Blank lines are ignored.
The current task adds comment lines beginning with # and reports malformed records
using an exported *LineError with Line (one-based physical line) and Text fields.
Load returns the valid prefix on malformed input or read failure. The caller's
reader errors remain inspectable. Process bytes returned alongside a read error;
EOF is normal completion. Successful errors are nil interfaces. No line-size cap
is promised. Empty keys and values are accepted. Readers are borrowed.
If malformed bytes arrive with an independent read failure, retain the valid
prefix and make both the structured parse error and reader cause inspectable.
