[Index](README.md) · [← Part 4 — The HTTP adapter](04-http-adapter.md)

# Part 5 — The gRPC adapter

Two files. Both contain a trap that is invisible until it bites, so read the explanations
rather than copying the code and moving on.

### `@GrpcAdvice` is the gRPC twin of `@RestControllerAdvice`

If Part 4's advice class made sense, this one will too — it is the same idea for a different
protocol.

Spring gRPC wraps every incoming call in an interceptor chain. When your RPC method throws,
the interceptor catches the exception and routes it to a matching handler method **before**
anything is written to the wire. So, exactly as in Part 4, no gRPC service method needs a
`try`/`catch`.

The annotation you write is `@GrpcAdvice`, and it is meta-annotated `@Component` — so Spring's
component scan finds it and registers it automatically.

### Why the handler annotation takes `Class[]`

```java
@GrpcExceptionHandler(DomainException.class)
```

The framework dispatches on **exception type**. Internally it builds a
`Map<Class<? extends Throwable>, Method>` and, when an exception arrives, walks up the type
hierarchy to find the closest match. That is why you name the exception class rather than
writing a conditional inside one big method.

> **Name collision — read this before you write the import.**
>
> There are two different types called `GrpcExceptionHandler` in Spring gRPC:
>
> | Type | Package | What it is |
> |---|---|---|
> | `GrpcExceptionHandler` | `org.springframework.grpc.server.advice` | **The annotation you write** |
> | `GrpcExceptionHandler` | `org.springframework.grpc.server.exception` | An internal interface |
>
> Identical simple names. IDE auto-import picks the wrong one roughly half the time. If your
> `@GrpcExceptionHandler` method "does not work" or the annotation appears unresolved, check
> which package your import came from — you almost certainly want `...server.advice`.

The handler method may return one of `Status`, `StatusException`, `StatusRuntimeException`, or
`Throwable`. We use `StatusException`.

### The projection is two steps, not one

This is the most important idea in the whole tutorial, and `ERROR_HANDLING.md` §3.2 is
explicit that it must not be simplified.

Step one is Part 1's rule — **category → gRPC code** — implemented here as `codeOf(category)`,
the gRPC adapter's own exhaustive `switch`. But `CONFLICT` is the exception:

| Conflict kind | gRPC | Retryable |
|---|---|---|
| `_CONCURRENT_MODIFICATION` (lost optimistic lock) | `ABORTED` (10) | **yes** |
| `_INVALID_TRANSITION` | `FAILED_PRECONDITION` (9) | no |
| `_ALREADY_EXISTS` | `FAILED_PRECONDITION` (9) | no |

All three are `CONFLICT`, and `CONFLICT`'s default is `FAILED_PRECONDITION`. So the category
alone cannot determine the gRPC code. A code whose status differs declares an **override** —
a table entry, keyed by the code string:

```
CONFLICT → FAILED_PRECONDITION (9)                     // category default
WORK_ORDER_CONCURRENT_MODIFICATION → ABORTED (10)      // code-level override
```

Precedence: **code override → category → status.**

> **An override, not a naming convention.** `_CONCURRENT_MODIFICATION` reads well but does
> nothing on its own — the code must be in the override table to project to `ABORTED`. A
> rename changes nothing; a *missing* override falls back to the safe `FAILED_PRECONDITION`
> default, never a retry storm.

Why it matters: `ABORTED` tells a gRPC client *"retry the whole transaction"*;
`FAILED_PRECONDITION` tells it *"this will never succeed"*. Conflate them and clients either
retry a permanent failure forever, or give up on a transient one.

The overrides are injected (a service can't be known to the adapter module). The service-side
bean is Part 6a; the merge is Part 6b.

### `GrpcStatusFactory` — building the status

Create `GrpcStatusFactory.java`. It is an **instance** class (not static methods) because it
carries the injected override table.

```java
package com.factoryos.common.error;

import java.util.Map;

import com.google.protobuf.Any;
import com.google.rpc.BadRequest;
import com.google.rpc.ErrorInfo;
import com.google.rpc.Status;

import io.grpc.StatusException;
import io.grpc.protobuf.StatusProto;

/**
 * Projects a {@link DomainException} onto a gRPC {@code google.rpc.Status}.
 *
 * <p>
 * Precedence (ERROR_HANDLING.md section 3.2): a code-level override wins if one is
 * registered, otherwise the category's default applies.
 */
public class GrpcStatusFactory {

    private static final String DOMAIN = "factoryos";

    /** Code-level overrides, keyed by the code string. Merged from provider beans. */
    private final Map<String, io.grpc.Status.Code> overrides;

    public GrpcStatusFactory(Map<String, io.grpc.Status.Code> overrides) {
        this.overrides = Map.copyOf(overrides);
    }

    /** Category → gRPC code. Exhaustive switch, no default: a new category fails to compile. */
    static io.grpc.Status.Code codeOf(ErrorCategory category) {
        return switch (category) {
            case MALFORMED_REQUEST, VALIDATION ->
                    io.grpc.Status.Code.INVALID_ARGUMENT;
            case AUTHENTICATION -> io.grpc.Status.Code.UNAUTHENTICATED;
            case AUTHORIZATION  -> io.grpc.Status.Code.PERMISSION_DENIED;
            case NOT_FOUND      -> io.grpc.Status.Code.NOT_FOUND;
            case CONFLICT       -> io.grpc.Status.Code.FAILED_PRECONDITION;
            case RATE_LIMIT     -> io.grpc.Status.Code.RESOURCE_EXHAUSTED;
            case DEPENDENCY     -> io.grpc.Status.Code.UNAVAILABLE;
            case INTERNAL       -> io.grpc.Status.Code.INTERNAL;
        };
    }

    public io.grpc.Status.Code toStatusCode(ErrorCode code) {
        // Code-level override wins if registered; otherwise the category default.
        io.grpc.Status.Code override = overrides.get(code.code());
        return override != null ? override : codeOf(code.category());
    }

    public StatusException toStatusException(DomainException ex) {
        io.grpc.Status.Code grpcCode = toStatusCode(ex.code());

        ErrorInfo.Builder info = ErrorInfo.newBuilder()
                .setReason(ex.code().code())
                .setDomain(DOMAIN)
                .putMetadata("category", ex.code().category().name())
                .putMetadata("retryable", Boolean.toString(ex.code().retryable()));

        Status.Builder status = Status.newBuilder()
                .setCode(grpcCode.value())
                .setMessage(ex.getMessage())
                .addDetails(Any.pack(info.build()));

        if (!ex.fieldErrors().isEmpty()) {
            BadRequest.Builder bad = BadRequest.newBuilder();
            for (DomainException.FieldError fe : ex.fieldErrors()) {
                BadRequest.FieldViolation.Builder v = BadRequest.FieldViolation.newBuilder()
                        .setField(fe.field());
                if (fe.message() != null) {
                    v.setDescription(fe.message());
                }
                bad.addFieldViolations(v);
            }
            status.addDetails(Any.pack(bad.build()));
        }

        // Packs the detail into the grpc-status-details-bin trailer. A bare Status
        // would give the right code with NO detail reaching the client.
        return StatusProto.toStatusException(status.build());
    }
}
```

### `GrpcStatusOverrideProvider` — how a service declares an override

The adapter module cannot import a service's enums, so a service contributes its overrides
through this one-method interface. Create `GrpcStatusOverrideProvider.java`:

```java
package com.factoryos.common.error;

import java.util.Map;

/** Lets a service contribute gRPC overrides for its own codes, keyed by code string. */
@FunctionalInterface
public interface GrpcStatusOverrideProvider {
    Map<String, io.grpc.Status.Code> overrides();
}
```

**Trap 1 — a `Status` cannot carry trailers.** Everything above exists to avoid this
mistake:

```java
// WRONG: correct status code, but ErrorInfo never reaches the client.
return Status.FAILED_PRECONDITION.asException();

// RIGHT: packs the google.rpc.Status into the grpc-status-details-bin trailer.
return StatusProto.toStatusException(status.build());
```

`ERROR_HANDLING.md` §5 requires a client to classify an error from the status code **plus
`ErrorInfo.reason`**. The first form satisfies the code and loses the reason — the error
arrives, but the client cannot tell *which* error it is. It compiles, it returns the right
number, and it silently breaks the contract. That is the worst kind of bug, which is why
Part 7 asserts the reason explicitly.

> **Two lookalike classes.** `com.google.rpc.StatusProto` contains **only descriptors** and
> has no `toStatusException`. The helper lives in **`io.grpc.protobuf.StatusProto`**. Mixing
> them up gives a confusing "cannot find symbol". Note that both are on the classpath, so
> your IDE will happily offer the wrong one.

**Trap 2 — the fallback is `UNKNOWN`, not `INTERNAL`.** If an exception matches no handler,
Spring gRPC calls `Status.fromThrowable(t)`, which walks the cause chain looking for a
`StatusException`, finds none, and produces `Status.UNKNOWN`. Verify it yourself — this is
what an unclassified exception becomes:

```
fallback for unhandled exception = UNKNOWN (value 2)
```

But `ERROR_HANDLING.md` §2 requires an unhandled failure to be `INTERNAL_ERROR` →
**`INTERNAL` (13)**. So without a catch-all, every unclassified bug reports the wrong code.
That is a correctness requirement, not a nicety.

### `GrpcErrorAdvice` — the adapter itself

Create `GrpcErrorAdvice.java`:

```java
package com.factoryos.common.error;

import org.springframework.grpc.server.advice.GrpcAdvice;
import org.springframework.grpc.server.advice.GrpcExceptionHandler;

import io.grpc.StatusException;

/**
 * The gRPC transport adapter: one place that projects every {@link DomainException}
 * onto a {@code google.rpc.Status} with typed detail.
 *
 * <p>
 * Note the import above: {@code org.springframework.grpc.server.advice} holds the
 * ANNOTATION. There is a same-named interface in
 * {@code org.springframework.grpc.server.exception} — importing that one instead
 * will not compile.
 */
@GrpcAdvice
public class GrpcErrorAdvice {

    private final GrpcStatusFactory statusFactory;

    /** The factory (with its merged override table) is injected as a bean — Part 6b. */
    public GrpcErrorAdvice(GrpcStatusFactory statusFactory) {
        this.statusFactory = statusFactory;
    }

    @GrpcExceptionHandler(DomainException.class)
    public StatusException handleDomainException(DomainException ex) {
        // MUST project from the code, not the class: both conflict kinds are
        // DomainException, and only a code-level override distinguishes ABORTED
        // from the FAILED_PRECONDITION default.
        return statusFactory.toStatusException(ex);
    }

    /**
     * Catch-all. Without this, an unclassified exception falls through to the
     * framework fallback, which yields Status.UNKNOWN (2) — but the standard
     * requires INTERNAL (13) for an unhandled failure.
     */
    @GrpcExceptionHandler(Exception.class)
    public StatusException handleUnexpected(Exception ex) {
        DomainException internal = new DomainException(
                CommonErrorCode.INTERNAL_ERROR,
                "An internal error occurred. Quote the traceId when contacting support.",
                java.util.List.of(), ex);
        return statusFactory.toStatusException(internal);
    }
}
```

**The catch-all is deliberate and required.** `@GrpcExceptionHandler(Exception.class)` claims
everything not otherwise matched, which is what turns the framework's default `UNKNOWN` (2)
into the standard's `INTERNAL` (13). The exception is attached as the **cause** of the new
`DomainException`, so the stack trace still reaches your logs even though the client sees only
a constant message — the same 5xx rule you applied in Part 4.

### What the client receives

For a work order that cannot move from `closed` to `released`, the client gets:

- gRPC status `FAILED_PRECONDITION` (9) — derived from category `CONFLICT`
- `ErrorInfo.reason` = `WORK_ORDER_INVALID_TRANSITION` — **the same string the REST response
  puts in `code`**
- `ErrorInfo` metadata: `category=CONFLICT`, `retryable=false`

That shared string is the payoff of the entire design. A client switches on `reason` and
behaves identically whether the request went over HTTP or gRPC. `ERROR_HANDLING.md` §5 states
the two rules that make this work:

> - **The detail is protobuf, not a formatted string.** A string message cannot be parsed
>   reliably across languages and breaks whenever the format changes.
> - **Use the standard `google.rpc` detail types**, not a bespoke message, so generic gRPC
>   tooling and proxies can render the error without knowing FactoryOS.

We use `ErrorInfo` and `BadRequest` — both standard types from `google.rpc`, which is why a Go
client can decode a Java service's error without any shared FactoryOS code.

### Verify it compiles

```bash
cd services && mvn -o -B test -pl ../libs/factoryos-common-error-adapters -am
```

### Check yourself

- Three new files exist: `GrpcStatusFactory.java`, `GrpcStatusOverrideProvider.java`, and
  `GrpcErrorAdvice.java`
- The import is `org.springframework.grpc.server.advice.GrpcExceptionHandler` (the annotation)
- You can explain why both `_INVALID_TRANSITION` and `_CONCURRENT_MODIFICATION` are
  `CONFLICT` yet map to different gRPC codes
- You can explain why the override is an explicit table entry, not a naming suffix
- You can explain what goes wrong if you return `Status.X.asException()` instead of using
  `StatusProto.toStatusException(...)`

---


---

[← Part 4 — The HTTP adapter](04-http-adapter.md) · [Index](README.md) · [Part 6a — Adopting it: the domain side →](06a-domain-adoption.md)
