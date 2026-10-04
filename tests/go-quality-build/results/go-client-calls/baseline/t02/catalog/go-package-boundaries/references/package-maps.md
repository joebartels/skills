# Illustrative package maps

Arrows mean Go imports. Each directory below represents a package; multiple files in that directory share its visibility and dependency boundary. Paths and project sizes illustrate choices, not thresholds or required names. Each example can live in one module. `internal` limits imports to the parent tree; `cmd` is an organizational convention.

## Small CLI: keep a cohesive operation together

```text
go.mod
cmd/clean/main.go             flags, process exit, invocation
internal/clean/clean.go       clean operation and validation
internal/clean/parse.go       related parsing helpers
internal/clean/files.go       local file reads/writes

cmd/clean -> internal/clean -> standard library
```

The local file representation and parser serve one command and change together. New files improve navigation without exported glue or package cycles. Adding a dry-run option is not evidence for separate parser, domain, port, and filesystem packages. A one-directory `package main` is also reasonable for a tiny command without reusable behavior; an implementation package pays for itself only when its use or cohesion warrants it.

## Small public library: preserve a coherent vocabulary

```text
go.mod
range.go                     Range and its operations
parse.go                     Parse producing Range
format.go                    formatting the same representation
range_test.go

external caller -> example.org/ranges -> standard library
```

Closely related types and functions remain in package `ranges`. A new parse function does not require a `parser` subpackage. Moving exported types would change consumers' import paths and type identity; do not perform that migration merely to match a directory map.

## Medium service: separate implementations when change pressure warrants it

Start with a feature package containing operation, HTTP handling, and storage in separate files. If they share ownership and evolution, that may remain appropriate after adding a scheduled caller. Reuse the operation without duplicating its validation or sequence.

When the transport, storage format, and outbound integration acquire independent maintenance or dependencies, separate their concrete implementations:

```text
cmd/service/main.go          configure and wire server and dependencies
cmd/worker/main.go           configure and wire background consumer
internal/orders/orders.go   reusable Place operation, invariants, sequencing
internal/orderhttp/http.go  HTTP input/status mapping
internal/orderjobs/jobs.go  queue decoding and operation invocation
internal/orderstore/sql.go  SQL schema and persistence implementation
internal/carrier/client.go  outbound carrier protocol

cmd/service -> orderhttp, orders, orderstore, carrier
cmd/worker  -> orderjobs, orders, orderstore, carrier
orderhttp, orderjobs -> orders
orderstore, carrier -> orders  (only if their method signatures need its types)
orders -> standard library   (does not import the concrete adapters)
```

`Place` validates and sequences persistence and dispatch for both callers. `orderhttp` is an ingress adapter; it is not the operation. `orderstore` owns SQL and record encoding. `carrier` owns the remote request representation. The commands acquire concrete resources, connect dependencies, and arrange their cleanup. Shared configuration can remain beside the commands unless meaningful reuse warrants another boundary.

A **port** is a consumer contract such as the narrow methods Place needs. An **adapter** is concrete translating code such as an SQL store or HTTP client. A needed port can live in `orders`; its implementation lives separately when independent evolution warrants that boundary. Some dependencies can be concrete values or functions without a named port. Neither adding an adapter nor adding a package mandates a new interface.

Alternative: keep `orders` and its ingress files together, separating only the independently changing SQL implementation. Split each responsibility for its own evidence; copying every box would overbuild the service.

## Large multi-feature application

```text
cmd/shop/main.go             composition and process lifecycle
internal/catalog/           catalog operations and types
internal/cataloghttp/       independently evolving catalog ingress
internal/catalogstore/      catalog persistence
internal/billing/           billing operations and types
internal/billingrpc/        billing RPC ingress
internal/billingjobs/       scheduled billing operations
internal/billingstore/      billing persistence
internal/paymentclient/     independently maintained external protocol
internal/moneywire/         real shared amount codec used by two integrations

cmd/shop -> ingress, jobs, operations, concrete implementations
cataloghttp, catalogstore -> catalog
billingrpc, billingjobs, billingstore -> billing
paymentclient -> billing   (if the adapter consumes billing types)
billingstore, paymentclient -> moneywire -> standard library
```

Feature ownership remains visible. Unrelated catalog and billing storage are not grouped into a central repository package just because both use SQL. The shared codec has actual consumers and no dependency on feature operations. If no such common protocol exists, keep each representation with its owner.

Separate worker binaries can be useful for distinct process lifetimes or deployment requirements. Separate modules need a release, ownership, or external reuse reason; more source files alone do not supply one. Avoid moving established public package paths casually.

## Component palette

| Component | Stay with caller / new file | Boundary evidence |
| --- | --- | --- |
| HTTP, RPC, CLI ingress | One small operation and cohesive translation | Independent protocol changes, transport dependencies, reusable operation with distinct ingress owners |
| Parsing and validation | Local syntax parsing; invariants with the operation | Genuinely reused codec or protocol; keep operation invariants available across ingress paths |
| SQL, file, cache persistence | Simple local representation sharing its operation's evolution | Independently maintained encoding/schema, backend dependencies, or ownership |
| Outbound API client | A small call with the same change path | Independent remote protocol, authentication or client reuse |
| Queues and workers | Scheduling wrapper near its caller | Independently changing message contract or runtime; invoke the same operation |
| Shared reusable code | Keep feature helpers with the feature | Multiple concrete consumers of one cohesive capability; avoid generic dumping grounds |
| Configuration and instrumentation | Process wiring near the entry point; local instrumentation near work | Shared host capability with real consumers; do not make core operations read process globals |

For every proposed split, identify the change it isolates and show the resulting imports. If the split mostly exports formerly private helpers or moves tightly shared invariants back and forth, separate files are often the better boundary.
