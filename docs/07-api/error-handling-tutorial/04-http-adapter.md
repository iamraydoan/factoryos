[Index](README.md) · [← Part 3 — The taxonomy module](03-taxonomy-module.md)

# Part 4 — The HTTP adapter

Now the domain layer gets projected onto HTTP. Three files, in
`libs/factoryos-common-error-adapters/src/main/java/com/factoryos/common/error/`.

Note the module change: this is `factoryos-common-error-adapters`, which **does** have
Spring and gRPC dependencies. The domain module from Part 3 stays clean; the transport
knowledge lives here.

### `@RestControllerAdvice`: what it actually does

Before using it, understand the mechanism — it is the whole reason no controller needs error
handling.

Spring MVC routes an incoming request to a controller method. If that method **throws**, the
exception would normally propagate out and become a container error page. `@RestControllerAdvice`
changes that: it registers a class as a **global exception interceptor**. When any controller
anywhere throws, Spring looks for a matching `@ExceptionHandler` method in the advice class
and calls *that* instead.

```
GET /api/v1/work-orders?limit=500
      ↓
WorkOrderController.listWorkOrders(...)
      ↓  throws DomainException(PAGE_SIZE_OUT_OF_RANGE)
ProblemDetailExceptionHandler.handleDomainException(ex, request)   ← intercepts
      ↓
HTTP 422  application/problem+json
```

The payoff: **one** class handles errors for **every** controller. This is what
`ERROR_HANDLING.md` §8 means by *"One per protocol, mapping category → status, written once
and reused by every handler."* You will not write a `try`/`catch` in a controller again.

> `@RestControllerAdvice` is `@ControllerAdvice` plus `@ResponseBody` — the latter means the
> returned object is serialized straight to the response body as JSON, rather than being
> treated as a view name.

### The wire format

`ERROR_HANDLING.md` §4 specifies **RFC 9457 Problem Details** with
`Content-Type: application/problem+json`. Here is the standard's own example:

```json
{
  "type": "https://factoryos.dev/errors/work-order-invalid-transition",
  "title": "Invalid state transition",
  "status": 409,
  "detail": "Work order cannot go from 'closed' to 'released'.",
  "instance": "/api/v1/work-orders/0192f3a1-7c4e-7b2a-9f01-3d5e8a1b2c34",
  "code": "WORK_ORDER_INVALID_TRANSITION",
  "retryable": false,
  "timestamp": "2026-09-15T08:30:00Z",
  "errors": [
    { "field": "state", "current": "closed", "allowed": ["released"] }
  ]
}
```

Spring provides `ProblemDetail` — a class representing exactly this shape, including the
RFC's `type`, `title`, `status`, `detail` and `instance` members. Anything beyond those (our
`code`, `retryable`, `traceId`, `errors[]`) is attached with `setProperty`.

### `ProblemDetailFactory` — building the payload

Create `ProblemDetailFactory.java`:

```java
package com.factoryos.common.error;

import java.net.URI;
import java.time.Instant;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import org.springframework.http.ProblemDetail;

/**
 * Projects a {@link DomainException} onto an RFC 9457 Problem Details payload.
 */
public final class ProblemDetailFactory {
    private ProblemDetailFactory() {
    }

    static int statusOf(ErrorCategory category) {
        return switch (category) {
            case MALFORMED_REQUEST -> 400;
            case VALIDATION -> 422;
            case AUTHENTICATION -> 401;
            case AUTHORIZATION -> 403;
            case NOT_FOUND -> 404;
            case CONFLICT -> 409;
            case RATE_LIMIT -> 429;
            case DEPENDENCY -> 503;
            case INTERNAL -> 500;
        };
    }

    public static ProblemDetail from(DomainException ex, String instancePath, String traceId) {
        ErrorCode code = ex.code();
        ErrorCategory category = code.category();
        ProblemDetail problem = ProblemDetail.forStatus(statusOf(category));
        problem.setType(URI.create(code.typeUri()));
        problem.setTitle(code.title());
        problem.setDetail(ex.getMessage());
        if (instancePath != null) {
            problem.setInstance(URI.create(instancePath));
        }
        problem.setProperty("code", code.code());
        problem.setProperty("retryable", code.retryable());
        if (traceId != null) {
            problem.setProperty("traceId", traceId);
        }
        problem.setProperty("timestamp", Instant.now().toString());
        List<Map<String, Object>> errors = ex.fieldErrors().stream()
                .map(fe -> {
                    Map<String, Object> entry = new LinkedHashMap<>();
                    entry.put("field", fe.field());
                    if (fe.message() != null) {
                        entry.put("message", fe.message());
                    }
                    if (fe.current() != null) {
                        entry.put("current", fe.current());
                    }
                    if (!fe.allowed().isEmpty()) {
                        entry.put("allowed", fe.allowed());
                    }
                    return entry;
                }).toList();
        if (!errors.isEmpty()) {
            problem.setProperty("errors", errors);
        }
        return problem;
    }
}
```

Three decisions worth understanding, all from `ERROR_HANDLING.md` §4:

**`ProblemDetail.forStatus(statusOf(category))` — the status comes from the HTTP adapter's
category projection.** Never from the exception. The category from Part 3 carries no status;
this adapter owns the HTTP projection, and `statusOf(...)` is that projection, written once
as an exhaustive `switch`. This is the projection from Part 1, happening for real.

**`setType(URI.create(code.typeUri()))` — derived, not stored.** `typeUri()` kebab-cases the
code, so `PAGE_SIZE_OUT_OF_RANGE` becomes
`https://factoryos.dev/errors/page-size-out-of-range`. Because it is computed from the code,
it cannot drift from it.

**`title` and `detail` are separate fields.** `title` comes from the code and is identical
for every instance of that code ("Page size out of range"). `detail` comes from the
exception and describes *this* occurrence ("...got 500"). The standard is explicit that a
single `message` field would collapse the two and leave clients unable to tell "what kind of
error" from "what happened this time".

### `TraceIdFilter` — correlating a report with a log line

Create `TraceIdFilter.java`:

```java
package com.factoryos.common.error;

import java.io.IOException;
import java.util.UUID;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;

import org.springframework.web.filter.OncePerRequestFilter;

/**
 * Assigns a correlation id to every request so a client report can be matched to a
 * server log entry. Honours an incoming {@code X-Trace-Id} if present.
 */
public class TraceIdFilter extends OncePerRequestFilter {

    public static final String TRACE_ID_HEADER = "X-Trace-Id";
    private static final ThreadLocal<String> CURRENT = new ThreadLocal<>();

    public static String currentTraceId() {
        return CURRENT.get();
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain chain)
            throws ServletException, IOException {
        String incoming = request.getHeader(TRACE_ID_HEADER);
        String traceId = incoming == null || incoming.isBlank() ? UUID.randomUUID().toString() : incoming;
        CURRENT.set(traceId);
        response.setHeader(TRACE_ID_HEADER, traceId);
        try {
            chain.doFilter(request, response);
        } finally {
            // Threads are reused; a stale id would leak into the next request.
            CURRENT.remove();
        }
    }
}
```

A **servlet filter** runs before and after every request. `OncePerRequestFilter` is Spring's
base class that guarantees it runs exactly once per request, even if the request is
internally forwarded.

Why the `ThreadLocal`: the trace id is needed deep inside error handling, but threading it
through every method signature would be invasive. A `ThreadLocal` gives each request its own
copy. **The `finally { CURRENT.remove(); }` is not optional** — servlet containers reuse
threads, so a stale id would leak into the *next* request on that thread. If you ever see a
trace id that belongs to somebody else's request, a missing `remove()` is the cause.

**Honouring an incoming `X-Trace-Id`** lets a client (or an upstream gateway) supply its own
id, which is how you correlate a single user action across several services.

### `ProblemDetailExceptionHandler` — the adapter itself

Create `ProblemDetailExceptionHandler.java`:

```java
package com.factoryos.common.error;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.slf4j.event.Level;
import org.springframework.http.ProblemDetail;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.context.request.WebRequest;

/**
 * The HTTP adapter: projects every {@link DomainException} onto RFC 9457 Problem
 * Details, so no controller needs its own error handling.
 */
@RestControllerAdvice
public class ProblemDetailExceptionHandler {

    private static final Logger log = LoggerFactory.getLogger(ProblemDetailExceptionHandler.class);

    /** Constant detail for 5xx — never echo an internal message to a client. */
    private static final String INTERNAL_DETAIL =
            "An internal error occurred. Quote the traceId when contacting support.";

    @ExceptionHandler(DomainException.class)
    public ResponseEntity<ProblemDetail> handleDomainException(DomainException ex, WebRequest request) {
        ErrorCode code = ex.code();
        String instance = request.getDescription(false).replaceFirst("^uri=", "");
        String traceId = TraceIdFilter.currentTraceId();
        log.atLevel(toSlf4jLevel(code.severity()))
                .setMessage("Domain error {code} on {instance}")
                .addKeyValue("code", code.code())
                .addKeyValue("category", code.category())
                .addKeyValue("traceId", traceId)
                .setCause(needsStackTrace(code.severity()) ? ex : null)
                .log();
        ProblemDetail problem = ProblemDetailFactory.from(ex, instance, traceId);
        return ResponseEntity.status(ProblemDetailFactory.statusOf(code.category())).body(problem);
    }

    /** Catch-all: unclassified failures are reported as INTERNAL_ERROR, never echoed. */
    @ExceptionHandler(Exception.class)
    public ResponseEntity<ProblemDetail> handleUnexpected(Exception ex, WebRequest request) {
        String instance = request.getDescription(false).replaceFirst("^uri=", "");
        String traceId = TraceIdFilter.currentTraceId();

        log.atError()
                .setMessage("Unhandled exception")
                .addKeyValue("instance", instance)
                .addKeyValue("traceId", traceId)
                .setCause(ex)
                .log();

        // The detail goes to the log; the client gets a constant plus the id.
        ProblemDetail problem = ProblemDetailFactory
                .from(new DomainException(CommonErrorCode.INTERNAL_ERROR, INTERNAL_DETAIL), instance, traceId);
        return ResponseEntity.status(ProblemDetailFactory.statusOf(CommonErrorCode.INTERNAL_ERROR.category()))
                .body(problem);
    }

    private static Level toSlf4jLevel(ErrorSeverity severity) {
        return switch (severity) {
            case INFO -> Level.INFO;
            case WARNING -> Level.WARN;
            case ERROR, CRITICAL -> Level.ERROR;
        };
    }

    private static boolean needsStackTrace(ErrorSeverity severity) {
        return severity == ErrorSeverity.ERROR || severity == ErrorSeverity.CRITICAL;
    }
}
```

The pieces that matter:

**`@ExceptionHandler(DomainException.class)`** — the handler for our classified errors. The
returned `ResponseEntity` carries the status derived from the category.

**`@ExceptionHandler(Exception.class)` — the catch-all.** Without it, anything that is *not*
a `DomainException` (a `NullPointerException`, a database timeout) escapes with Spring's
default body, which leaks internals. This is the safety net, and it is what fixes defect #1
and #3 from Part 0.

**Log level comes from `code.severity()`, not from a decision at this call site.** A bad
cursor logs at INFO; a downstream outage logs at ERROR with a stack trace. This is
`ERROR_HANDLING.md` §7 in code — the severity was declared once on the code, and this
handler just obeys.

**The 5xx rule.** `ERROR_HANDLING.md` §4 is explicit:

> **Never echo an internal error message on a 5xx.** Unhandled exceptions carry variable
> names, table and column names, or absolute file paths. Return a constant `detail` plus a
> `traceId`; put the detail in the log.

That is why `INTERNAL_DETAIL` is a constant string. The *real* message and the stack trace go
to `log.error(...)`, where an engineer can read them; the client gets a stable sentence and
an id to quote. A `NullPointerException` message often names the exact field and class that
broke — useful to you, a gift to an attacker.

### Verify it compiles

```bash
cd services && mvn -o -B test -pl ../libs/factoryos-common-error-adapters -am
```

**The `-am` flag is required here.** This module depends on `factoryos-common-error`, and
without `-am` Maven does not build that dependency first, so the build fails with:

```
Could not find artifact com.factoryos:factoryos-common-error:jar:0.1.0-SNAPSHOT
```

If you see that message, you forgot `-am`. It is not a problem with your code.

### Check yourself

- Three new files exist in `libs/factoryos-common-error-adapters/.../common/error/`
- `mvn -o -B test -pl ../libs/factoryos-common-error-adapters -am` reports `BUILD SUCCESS`
- You can explain why the handler never echoes `ex.getMessage()` for a 5xx
- You can explain why `TraceIdFilter` calls `CURRENT.remove()` in a `finally` block

---


---

[← Part 3 — The taxonomy module](03-taxonomy-module.md) · [Index](README.md) · [Part 5 — The gRPC adapter →](05-grpc-adapter.md)
