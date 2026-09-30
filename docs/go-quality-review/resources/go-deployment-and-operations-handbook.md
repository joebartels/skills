# Go Deployment and Operations Handbook
**Section 7: Deployment, Containerization, Health Probes, and Runtime Operations**

---

## Section 1: Containerization & Build Optimization

Modern Go microservices require reproducible, minimal, and secure containerization artifacts. Build strategies must leverage Docker multi-stage builds, intelligent layer caching, and compiler flags to optimize binary footprint and security posture.

* **Multi-Stage Builds:** Use an official Go toolchain image (e.g., `golang:1.24-bookworm`) exclusively in the compilation stage. Copy only the final statically compiled binary artifact into a minimal, clean runtime image (e.g., `scratch` or `distroless/static`). This isolates compiler binaries, Go SDK source tools, and build-time header files from production deployment targets.
* **Layer Caching Mechanics:** Structure Dockerfiles to maximize layer reuse during incremental CI/CD pipelines. Copy `go.mod` and `go.sum` first and run `RUN go mod download` before copying application source code. This ensures dependency downloads are cached unless module manifests change.
* **Binary Optimization & Reproducibility:** Build binaries with strict flags: `CGO_ENABLED=0` (statically links all Go runtime packages, eliminating glibc/musl dependencies), `-trimpath` (removes host filesystem paths from stack traces for security and reproducible builds), and `-ldflags="-s -w"` (strips DWARF debugging symbols and symbol tables to reduce binary size by 30–40%).

```dockerfile
# Idiomatic Multi-Stage Dockerfile Strategy
FROM golang:1.24-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/app ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/app /app
USER 10001:10001
ENTRYPOINT ["/app"]
```

## Section 2: Base Images, Privileges & Secret Isolation

Container security relies on minimal attack surfaces, strict non-root user execution, and absolute isolation of credentials during build time.

* **Base Image Selection:** Default to `scratch` or `gcr.io/distroless/static-debian12` for pure Go services. Use `alpine` or minimal Debian images strictly when operational requirements explicitly mandate C-library wrappers (e.g., CGO for SQLite/Kafka bindings) or bundled diagnostic tools.
* **Non-Root Privilege Enforcement:** Explicitly define an unprivileged numeric user and group ID (e.g., `USER 10001:10001`) in the Dockerfile. Never permit containerized binaries to run as root (UID 0), preventing container-escape privilege escalation vulnerabilities.
* **Build Secret Isolation:** Exclude sensitive credentials, local `.env` files, and test coverage dumps via `.dockerignore`. NEVER pass private module access tokens or credentials via Docker `ARG` or `ENV` instructions (which leak keys into immutable image layer metadata). Utilize BuildKit's `RUN --mount=type=secret,id=netrc` for secure credential mounting during module fetching.

> **CONTAINER HARDENING MANDATE**
> **Security Rule:** Hardcoding tokens in Dockerfile ARG/ENV instructions leaves credentials permanently readable in 'docker inspect' and registry layer manifests. Always use BuildKit secret mounts or ephemeral runtime injection.

## Section 3: Health Probe Architecture & Orchestration

Orchestrators (Kubernetes, ECS) require clear, decoupled application signals to distinguish process responsiveness from dependency availability.

* **Probe Decoupling:** Explicitly separate Liveness (`/livez`), Readiness (`/readyz`), and Startup (`/startupz`) endpoints to match orchestrator lifecycle semantics.
* **Liveness Checks (Process Responsiveness):** Keep `/livez` strictly local to the process event loop. NEVER ping external databases, message brokers, or upstream HTTP services inside liveness checks. An upstream database outage must never cause local Liveness failures, which trigger cluster-wide cascading pod reboot loops.
* **Readiness & Startup Checks (Traffic Routing):** Wire `/readyz` checks to active dependency health (e.g., `db.PingContext(ctx)`). If a database pool degrades, `/readyz` returns HTTP 503 so load balancers safely stop routing traffic to the pod. Use `/startupz` to shield slow-starting services (e.g., cache warm-up) from premature liveness kills.

```go
// Decoupled Health Probe Handlers
mux.HandleFunc("GET /livez", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK) // Pure process responsiveness check
})

mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
    if isShuttingDown.Load() {
        http.Error(w, "shutting down", http.StatusServiceUnavailable)
        return
    }
    if err := db.PingContext(r.Context()); err != nil {
        http.Error(w, "database degraded", http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
})
```

## Section 4: Lifecycle Governance & Graceful Shutdown

Zero-downtime deployments require deterministic process teardown protocols synchronized with cluster ingress updates.

* **Termination Budget Alignment:** Configure container orchestrator termination budgets (e.g., Kubernetes `terminationGracePeriodSeconds: 45`) to exceed the application's combined ingress drain pause and shutdown timeout budget (e.g., 5s drain + 30s context timeout).
* **Signal Handling & Double-Interrupt Guard:** Bind `SIGTERM` and `SIGINT` signals to root context cancellation using `signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)`. Immediately invoke `stop()` after signal receipt so a second Ctrl+C forces an unblocked hard termination.
* **Ingress Propagation & Teardown Protocol:** Upon receiving `SIGTERM`: (1) Immediately flip `/readyz` to HTTP 503, (2) Pause for 3–5 seconds to allow cluster ingress controllers to update endpoints and stop routing new requests, (3) Execute `server.Shutdown(ctx)` to drain active connections, and (4) Close database pools and flush telemetry log buffers in reverse order of initialization.

## Section 5: Configuration & Secret Ingestion

Application binaries must remain immutable across environments, receiving environment configuration and secrets through secure runtime mechanisms.

* **Twelve-Factor Configuration:** Inject environment-specific parameters via environment variables or mounted configuration files. Never compile environment-specific endpoints, feature flags, or timeout values into application binaries.
* **Secure Secret Ingestion:** Mount sensitive credentials via ephemeral tmpfs RAM volumes (Kubernetes Secret volumes) or fetch credentials dynamically via Secret Manager SDKs (1Password, AWS Secrets Manager, Vault). Avoid writing unencrypted secrets to persistent container disks.
* **Fail-Fast Startup Validation:** Parse and validate all configuration fields at startup using structural validation rules (e.g., `envconfig` or `go-playground/validator`). If required parameters are missing or invalid, log a fatal error and terminate immediately before opening network listeners.

## Section 6: Production Operations & Cgroup Runtime Awareness

Production Go deployments require runtime visibility, cgroup-aware memory management, and standardized logging egress.

* **Diagnostic & Profiling Boundaries:** Expose dynamic log level toggles (`slog.LevelVar`) and profiling endpoints (`net/http/pprof`) exclusively behind internal administrative network listeners or authenticated management ports—never expose `/debug/pprof` to public ingress.
* **Cgroup-Aware Memory Limits (GOMEMLIMIT):** Set `GOMEMLIMIT` to 80–90% of the container memory limit to prevent Linux cgroup Out-Of-Memory (OOM) kills. Use community libraries like `automemlimit` to read cgroup limits dynamically at startup and configure `GOMEMLIMIT` automatically.
* **Telemetry Egress:** Write structured JSON logs exclusively to `os.Stdout` or `os.Stderr` to integrate natively with container log drivers (Fluentbit, Datadog, Vector) and cloud aggregators.

## Section 7: Reviewer Quick-Reference Grading Matrix

| Category | Review Standard | Anti-Pattern / Smell Indicator | Idiomatic Remediation |
| :--- | :--- | :--- | :--- |
| **Build Strategy** | Multi-stage Dockerfile with clean compilation isolation | Single-stage Dockerfile containing Go SDK and source files in production | Use `golang:1.24-bookworm AS builder` and copy static binary to `distroless/static` |
| **Binary Flags** | Static build with stripped symbols and clean paths | Default build with embedded host paths and DWARF tables | Build with `CGO_ENABLED=0 -trimpath -ldflags='-s -w'` |
| **Base Image** | Minimal runtime image (`scratch`/`distroless`) by default | Full Ubuntu or Debian image used without operational need | Deploy to `gcr.io/distroless/static-debian12:nonroot` |
| **Privileges** | Container executes under unprivileged numeric UID | Binary runs as root (`USER root` or omitted `USER` directive) | Specify `USER 10001:10001` in Dockerfile |
| **Build Secrets** | BuildKit secret mounts for dependency credentials | Private tokens passed in `ARG` or `ENV` instructions | Use `RUN --mount=type=secret,id=netrc` for `go mod download` |
| **Liveness Probe** | Probe checks process loop responsiveness locally | `/livez` handler pings external database or API | Keep `/livez` pure (return 200 OK); move DB checks to `/readyz` |
| **Readiness Probe** | Probe evaluates active dependency health and shutdown status | `/readyz` returns 200 OK without checking DB connectivity | Verify `db.PingContext(ctx)` and return 503 if shutting down |
| **Shutdown Flow** | 503 flip + 3-5s ingress drain pause before `server.Shutdown` | Immediate `server.Close()` on `SIGTERM` causing dropped requests | Use `signal.NotifyContext`, flip `/readyz`, pause 5s, call `server.Shutdown(ctx)` |
| **Memory Limits** | `GOMEMLIMIT` set to 80-90% of cgroup limit | Unset `GOMEMLIMIT` causing cgroup OOM kills before GC triggers | Set `GOMEMLIMIT` or import `KimMachineGun/automemlimit` |
| **Diagnostics** | `pprof` and admin endpoints isolated behind internal port | `/debug/pprof` exposed publicly on primary HTTP router | Mount `pprof` on dedicated internal admin port (e.g., `:9090`) |
