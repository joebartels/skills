# Design decisions for Go architecture reviews

Read the relevant sections when a judgment is contested or context-sensitive. These are decision aids for this skill, not requirements for every project. Verify version-sensitive behavior against primary documentation, the project's declared support, and its actual build/runtime configuration.

## Packages and layers

- Judge dependency direction by actual coupling. Transport or persistence details embedded in reusable business code can make that code hard to reuse or change. A small command or service can reasonably keep HTTP, SQL, and orchestration in one cohesive package. Require a separate domain, application, ports, or adapters package only when the separation solves a demonstrated problem.
- Package by responsibility and reader-facing API. Feature grouping can help when related behavior changes together; technical grouping can also work when it represents a coherent reusable capability. Import cycles are a compiler error, but the absence of a cycle does not prove good boundaries. Conversely, an unused layer adds navigation and mapping cost.
- A struct tag does not alone prove architectural leakage. Determine whether the type is a persistence record, a transport DTO, or a reusable domain type and whether consumers now depend on an unwanted schema. Add mapping where contracts diverge or independent evolution has value; avoid duplicate models that merely copy fields without a benefit.

See the Go team's [Organizing Go code](https://go.dev/blog/organizing-go-code), [Package names](https://go.dev/blog/package-names), and [module layout](https://go.dev/doc/modules/layout).

## Interfaces and construction

- Prefer the smallest interface needed by a real consumer, and consider concrete types before inventing an interface only for mocks. A producer-owned interface is appropriate when the package offers a protocol for multiple implementations, as with `io.Reader` or `hash.Hash`. Returning an interface can also be appropriate when the implementation is deliberately hidden or compatibility requires it; `New() *T` is not a universal constructor rule.
- Interface size is contextual: a cohesive three-method protocol may be better than three near-duplicate one-method interfaces. A compile-time implementation assertion is useful when it checks an intended external contract, but is not mandatory when assignments already make conformance clear.
- Inject dependencies when hidden coupling causes configuration, testing, concurrency, or lifecycle problems. A zero-value usable type, simple function, or package constant may need no constructor. Package `init` is not inherently wrong; hidden network I/O, unhandled failures, and work that cannot be shut down are the real concerns. Functional options help some public APIs but can obscure a small fixed set of required inputs.

See [Google Go Style Best Practices: interfaces](https://google.github.io/styleguide/go/best-practices.html#interfaces) and [Effective Go: interfaces](https://go.dev/doc/effective_go#interfaces_and_types) for the distinction between exposed and hidden contracts. Effective Go explains core language features but is not actively updated; verify version-specific advice against current documentation.

## Context, errors, and lifecycle as contracts

- For operations that can block or outlive a caller, check whether cancellation and deadlines reach the work. A context is usually a first argument to the call that needs it. Do not demand a context on every pure function. Storing a context in a long-lived struct usually hides per-call lifetime; scoped request-like objects and compatibility constraints can be exceptions.
- `QueryContext` and related APIs make cancellation possible, but the actual stopping behavior depends on the driver and operation. Avoid claiming that a particular backend query is guaranteed to stop merely because a context was passed. Choose time budgets from the caller's contract and operational constraints, not a universal duration.
- Translate errors at a boundary when callers need a stable meaning or a protocol response. Preserve useful low-level causes with `%w` only if exposing them through `errors.Is`/`As` is an intended API contract. A domain sentinel is one option, not a requirement for every missing row. Do not expose internal details to untrusted clients.
- Trace asynchronous work to an owner. Ask what success promises, how errors surface, how work is canceled or drained, and whether losing work is acceptable. A `WaitGroup`, `errgroup`, queue, or synchronous call can each be right. A service may need signal handling and shutdown; a library should generally expose lifecycle controls to its caller instead of owning process signals. Do not prescribe a fixed readiness delay.

See the Go team's [Contexts and structs](https://go.dev/blog/context-and-structs), [Canceling database operations](https://go.dev/doc/database/cancel-operations), and [Working with errors in Go 1.13](https://go.dev/blog/go1.13-errors).

- A library must make process-wide effects part of an explicit contract. Installing signal handlers, calling `os.Exit`, mutating the default HTTP mux, or replacing global logging from a constructor can interfere with unrelated host components. Expose lifecycle/configuration controls to the host when the library does not own the process; judge the demonstrated interference rather than banning every global registration.

## Review boundary examples

- A library defining a documented `Store` interface with multiple implementations may deserve praise for an explicit protocol. Its producer ownership is not a defect. Assess missing semantics only when the public contract truly omits behavior callers require; do not infer unbounded blobs or a performance problem without evidence.
- In a small HTTP service, a handler using SQL directly is not automatically an architectural defect. Hidden global DB initialization, discarded initialization errors, unowned receipt delivery, and raw database error text in a client response can be: they obscure dependencies, misrepresent success, or expose an accidental protocol contract. Suggest the smallest ownership change that fixes those effects before proposing a new package hierarchy or durable outbox. A zero-row update may intentionally be idempotent; check the contract before treating it as a defect.
- If a change only adjusts a local calculation without touching an API, dependency, boundary, or lifecycle contract, architecture may be Not applicable. A partial diff that changes a public interface but omits all callers may be Insufficient evidence if compatibility cannot be established.
