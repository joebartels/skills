# Checkout service changeset

The change adds `service/main.go`'s import of `example.com/checkout/shared/format` and adds the repository-root `go.work`. The sibling `shared` module is available in the development repository. The `service/go.mod` file was not changed.

Developers run `go test ./...` from `service` with the repository workspace active. The published release process packages only the `service` directory as a standalone module and builds it with `GOWORK=off` and `-mod=readonly`. This is the only affected release target. The intended minimum Go version is 1.24. No external module proxy serves `example.com/checkout/shared` yet. Review only the dependency and reproducibility consequences of this change.
