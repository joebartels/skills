# Development

The workspace contains the published SDK in `lib` and the support team's `streamctl` command in `tools/streamctl`. Run `make test` from the repository root; `make build` builds the command. Both modules support Go 1.22 and later.

The `registry` directory is a read-only mirror of upstream source releases used by the offline development environment. Go's module proxy is configured by the environment; normal Go commands can resolve the mirrored modules without network access.
