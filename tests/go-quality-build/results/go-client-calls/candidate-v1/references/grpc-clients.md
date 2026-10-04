# Unary gRPC clients

Read the pinned grpc-go version, generated service/client, resolver,
interceptors, credentials and effective service config. Reuse a `ClientConn`
and generated client under the host's lifecycle. Forward supplied options and
metadata without silently replacing credential, authority, resolver or retry
policy. Production credentials must fit the trust boundary; insecure local
test transports are not a production default. Check dial APIs against the
selected version rather than raising its Go/library minimum for convenience.

## Use the effective native policy

Inspect the policy actually selected for the method before adding a loop or
retry interceptor. Match service/method scope intentionally. Native policy
includes attempt count, eligible status codes and backoff; `maxAttempts` includes
the first configured attempt. Transparent retries are a separate mechanism and
may not consume that count. Applicable library caps, retry throttling, server
pushback and buffered request limits can further affect actual behavior.

In grpc-go, `WithDefaultServiceConfig` is a fallback when the resolver supplies
no valid service config. `WithDisableServiceConfig` changes that precedence.
Keep host/resolver policy when the contract calls for a fallback; do not disable
resolver configs merely to force your defaults. A valid empty resolver policy
can also supersede the fallback. Inspect the public effective method config
and actual local RPC executions when testing precedence.

Call commitment stops native configured replay. Receiving response headers or
exceeding the retry buffer can commit a call even when it later fails with a
retry-eligible status. An application retry loop can bypass that protection and
multiply native attempts. Use one generated invocation when native retries own
the policy. grpc-go's `WithDisableRetry` leaves transparent retries enabled;
do not infer Go behavior from another language's generic guide. Verify these
details against the selected implementation when versions change.

## Interpret the operation's outcome

`UNAVAILABLE` is a candidate transient, not proof that repeating a mutation is
safe. `DEADLINE_EXCEEDED` can accompany a completed server-side effect. `ABORTED`
may require restarting the enclosing transaction; `FAILED_PRECONDITION` may
require repairing state. Authentication/permission/configuration errors need
their own policy. Preserve promised `status.Code` classification through error
wrapping and retain uncertain completion instead of reporting no effect.

For server deduplication, pass the stable caller operation identity through the
agreed metadata field and preserve existing outgoing metadata. Copy metadata
before mutation. Use identical logical request data for configured retries and
subsequent caller invocations of that operation; don't generate a new identity
inside every call. A metadata label alone creates no server contract.

Give the unary operation one total context deadline, including native waiting
and retries. Enable wait-for-ready only if queuing during unavailable
connectivity fits that budget and the caller contract; it establishes neither
replay safety nor guaranteed completion. Set workload-appropriate message limits
and inspect any retry-buffer memory implications. Distinguish connectivity
state, health RPC outcomes and an application operation's actual success.

Verify the generated unary boundary with actual server execution counts,
metadata/payload/effects, resolver policy, headers before failure, status
classification, deadlines and borrowed connection reuse. Header commitment and
application-invocation counts cannot be demonstrated by a mocked retry helper.

Sources: [grpc-go default service configuration](https://pkg.go.dev/google.golang.org/grpc#WithDefaultServiceConfig),
[disabling retry](https://pkg.go.dev/google.golang.org/grpc#WithDisableRetry),
[native retries](https://grpc.io/docs/guides/retry/),
[status codes](https://grpc.io/docs/guides/status-codes/), and
[wait-for-ready](https://grpc.io/docs/guides/wait-for-ready/).
