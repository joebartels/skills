# joblist

joblist 1.x is a command used by shell scripts and a dashboard. Run it with
one JSON file argument. Input is a JSON array of objects with string id and
state fields. A JSON null document represents an empty input. Unknown fields
are ignored. The file must contain one complete JSON document.

The stable stdout format is one JSON object and a newline. It has a jobs
field containing the input records in their original order; records have id
and state fields even when their values are empty. An empty result has
"jobs":null. This distinction is used by the dashboard to show its empty
state. On success exit 0; on usage, read or input errors exit 2, emit no stdout,
and write a diagnostic to stderr. Diagnostic wording is not stable.

All Go code is private to this binary. Go 1.22 is supported.
