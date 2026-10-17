# Distribution

Codes supports Go 1.22.0 and later, with no nonstandard build tools or network
access needed for this module. The library and CLI are also consumed directly
from Go source releases: `go test ./...` and `go build ./cmd/codes` are supported
from the exported source tree. Published source builds do not run the catalog
editor or preview server.

`formats.csv` is the catalog maintained by the import team. `make preview`
refreshes the validation table and JSON schemas used by the form designers'
local preview server. `preview/` contains that server's disposable working
files. Run `make test` for development checks and `make source` to assemble the
working source export in `dist/codes-src.tar.gz`; its inventory follows Git's
source files, including new unignored files. Release CI builds that export.
