[Index](README.md) · [← Part 2 — Maven ground floor](02-maven-ground-floor.md)

# Part 3 — The taxonomy module

This part builds the domain layer. It is the longest part, because it creates six files —
but each one is small, and none of them import anything from Spring or gRPC.

### Where the files go

```
libs/factoryos-common-error/src/main/java/com/factoryos/common/error/
```

That directory tree already exists. The module is wired into the reactor and its `pom.xml`
is complete — it is waiting for these six classes.

### The constraint that shapes everything

Before writing any code, understand the rule this module enforces.

Open `libs/factoryos-common-error/pom.xml` and read its description:

> Language-neutral error taxonomy: categories, codes, and the domain error base.
> **Deliberately depends on NO transport library (no Spring, no gRPC)** so that
> ERROR_HANDLING.md section 1 is enforced by the compiler rather than by review.

The module has no dependencies beyond JUnit for tests. That is not an accident of
packaging — it is the design. `ERROR_HANDLING.md` §1 says the domain layer must not know
which protocol delivered the request. Here, that rule is not a convention you have to
remember while coding. **It is physically impossible to violate:** there is no
`HttpStatus` class on the classpath to reference, and no `io.grpc.Status` either.

Keep that in mind as you read the code below. The category carries no transport code at
all — not even as a plain `int` — because the numeric ranges `400–500` and `1–16` are
still HTTP and gRPC vocabulary, and they belong in the adapters, not here.

### `ErrorCategory` — the semantic class of a failure

Create `ErrorCategory.java`:

```java
package com.factoryos.common.error;

public enum ErrorCategory {
    MALFORMED_REQUEST,
    VALIDATION,
    AUTHENTICATION,
    AUTHORIZATION,
    NOT_FOUND,
    CONFLICT,
    RATE_LIMIT,
    DEPENDENCY,
    INTERNAL
}
```

**Why the mapping does *not* live here.** The tempting shortcut is to hang the two transport
codes off each constant — `MALFORMED_REQUEST(400, 3)` — and read them back later. We are not doing
that, and the reasons are the whole point of the design:

- **The domain would still know transport vocabulary.** `int` instead of `HttpStatus` makes
  the module transport-*dependency*-free, not transport-*knowledge*-free. `400` only means
  something as an HTTP status and `3` only as a gRPC code; both are protocol facts sitting in
  the domain layer.
- **`ERROR_HANDLING.md` §1.1 says a domain error carries no transport status.** A pair of
  transport codes on the category is exactly that status, in disguise. The category is the
  *input* to the projection; it is not the projection.
- **The mapping would not be exhaustive by construction any more.** §1.4 requires that adding
  a category force *both* adapters to be updated. A constructor argument makes every
  constant self-complete — the compiler stops being the guarantee. A `switch` in each adapter
  *without* a `default` branch restores it: a new category fails to compile in every adapter
  until each one has a case.
- **Adding a transport would change the domain.** A third protocol (SSE, a CLI, a new RPC
  style) would need a third value on the enum — a domain edit for a transport decision. With
  the mapping in the adapter, a new transport is a new adapter and nothing else.

So the category is just the vocabulary. The HTTP projection lives in the HTTP adapter
(Part 4); the gRPC projection lives in the gRPC adapter (Part 5).

> **`MALFORMED_REQUEST` projects to HTTP 400; `VALIDATION` projects to 422.** That
> split is intentional. A request that will not parse (`MALFORMED_REQUEST`) is *malformed
> input* → 400. A body that parses but carries invalid data
> (`VALIDATION`) is *unprocessable* → 422. Both project to gRPC `INVALID_ARGUMENT` (3),
> so gRPC clients see no difference. A bad cursor is malformed input too, so its codes
> (`INVALID_CURSOR`, `UNSUPPORTED_CURSOR_VERSION`, `CURSOR_SORT_KEY_MISMATCH`) belong to
> `MALFORMED_REQUEST` — they are distinguished by code, not category.

### `ErrorSeverity` — how loudly to log

Create `ErrorSeverity.java`:

```java
package com.factoryos.common.error;

public enum ErrorSeverity {
    INFO,
    WARNING,
    ERROR,
    CRITICAL
}
```
**Why severity belongs to the code, not the call site.** This is subtle and worth the
minute. A bad pagination cursor and a broken database connection are both, mechanically,
"the request failed". But only one of them should wake an on-call engineer at 3am.

If the log level were chosen where the exception is thrown, every developer would have to
make that judgement correctly, every time, forever. Instead the *code* declares its
severity once — `INVALID_CURSOR` is `INFO`, `DEPENDENCY_UNAVAILABLE` is `ERROR` — and the
logging adapter reads it. `ERROR_HANDLING.md` §7 maps each severity to a level:

| Severity | Level | Use |
|---|---|---|
| `INFO` | INFO | Expected client behaviour — a bad cursor, a 404 |
| `WARNING` | WARN | Worth attention — a conflict, a denied permission |
| `ERROR` | ERROR + stack trace | Service-level failure — a downstream outage |
| `CRITICAL` | ERROR + stack trace | Availability or data-integrity impact. Alert. |

### `ErrorCode` — the interface

Create `ErrorCode.java`. Before you read it, know **why it is an interface and not an
abstract class** — this is the Java fact that forces the whole design.

Java enums **cannot extend a class**, but they **can implement an interface**. The shared
codes and each domain's codes need to be separate enums (so a service does not carry every
other domain's vocabulary), and they must be interchangeable wherever a code is expected.
An interface is the only thing that satisfies both.

```java
package com.factoryos.common.error;

import java.util.Locale;

public interface ErrorCode {

    default String code() {
        if (this instanceof Enum<?> constant) {
            return constant.name();
        }
        throw new IllegalStateException(
                "ErrorCode must be an enum constant to derive its code: " + getClass().getName());
    }

    ErrorCategory category();

    String title();

    ErrorSeverity severity();

    boolean retryable();

    default String typeUri() {
        String kebab = code().toLowerCase(Locale.ROOT).replace('_', '-');
        return "https://factoryos.dev/errors/" + kebab;
    }
}
```

Two things to notice:

**`code()` is a `default` method.** Every enum constant already *is* the code —
`PAGE_SIZE_OUT_OF_RANGE` is spelled exactly as the contract spells it. Deriving it from
`Enum.name()` means no enum can misspell its own code, and you never write the string
twice. (If you implemented this as an abstract method instead, every enum would have to
restate a name it already has — and a typo would produce a code the contract does not
recognise.)

**`typeUri()` is derived, never stored.** `ERROR_HANDLING.md` §4 requires the `type` field
to be `https://factoryos.dev/errors/<kebab-case-code>`. Deriving it means it **cannot drift**
from the code it identifies — there is no second field to forget to update.

### `CommonErrorCode` — the shared vocabulary

Create `CommonErrorCode.java`. These are the codes every service can emit, resolvable
without any domain context (`ERROR_HANDLING.md` §6):

```java
package com.factoryos.common.error;

public enum CommonErrorCode implements ErrorCode {
    INTERNAL_ERROR(ErrorCategory.INTERNAL, "Unhandled failure", ErrorSeverity.CRITICAL, false),
    DEPENDENCY_UNAVAILABLE(ErrorCategory.DEPENDENCY, "Downstream dependency unavailable", ErrorSeverity.ERROR, true),
    RESOURCE_EXHAUSTED(ErrorCategory.RATE_LIMIT, "Quota exceeded", ErrorSeverity.WARNING, true),
    UNAUTHENTICATED(ErrorCategory.AUTHENTICATION, "Credentials missing or rejected", ErrorSeverity.WARNING, false),
    PERMISSION_DENIED(ErrorCategory.AUTHORIZATION, "Not permitted", ErrorSeverity.WARNING, false),
    MALFORMED_REQUEST(ErrorCategory.MALFORMED_REQUEST, "Malformed request", ErrorSeverity.INFO, false),
    MISSING_REQUIRED_FIELD(ErrorCategory.VALIDATION, "Required field missing", ErrorSeverity.INFO, false),
    INVALID_CURSOR(ErrorCategory.MALFORMED_REQUEST, "Invalid cursor", ErrorSeverity.INFO, false),
    UNSUPPORTED_CURSOR_VERSION(ErrorCategory.MALFORMED_REQUEST, "Unsupported cursor version", ErrorSeverity.INFO, false),
    CURSOR_SORT_KEY_MISMATCH(ErrorCategory.MALFORMED_REQUEST, "Cursor sort key mismatch", ErrorSeverity.INFO, false),
    PAGE_SIZE_OUT_OF_RANGE(ErrorCategory.VALIDATION, "Page size out of range", ErrorSeverity.INFO, false);

    private final ErrorCategory category;
    private final String title;
    private final ErrorSeverity severity;
    private final boolean retryable;

    private CommonErrorCode(ErrorCategory category, String title, ErrorSeverity severity, boolean retryable) {
        this.category = category;
        this.title = title;
        this.severity = severity;
        this.retryable = retryable;
    }

    @Override public ErrorCategory category() { return category; }
    @Override public String title() { return title; }
    @Override public ErrorSeverity severity() { return severity; }
    @Override public boolean retryable() { return retryable; }
}
```

Note that `code()` is **not** overridden. Each constant's name *is* its code, so the
interface's default method supplies it.

### `ErrorKey` — shared field names

Create `ErrorKey.java`. This class is named explicitly in the contract —
`errors.yaml` says to *"use the constants in `com.factoryos.common.error.ErrorKey` so
client and server cannot disagree on spelling."*

```java
package com.factoryos.common.error;

public final class ErrorKey {
    public static final String STATE = "state";
    public static final String LIMIT = "limit";
    public static final String CURSOR = "cursor";
    public static final String PAGE = "page";
    public static final String ID = "id";
    public static final String WORK_CENTER_ID = "workCenterId";

    private ErrorKey() {
    }
}
```

The private constructor is deliberate: this is a holder for constants, and it should never
be instantiated.

### `DomainException` — the base error

Create `DomainException.java`. This is the class every classified failure in FactoryOS
extends.

```java
package com.factoryos.common.error;

import java.util.List;
import java.util.Objects;

public class DomainException extends RuntimeException {

    private final ErrorCode code;
    private final transient List<FieldError> fieldErrors;

    public DomainException(ErrorCode code, String detail) {
        this(code, detail, List.of(), null);
    }

    public DomainException(ErrorCode code, String detail, List<FieldError> fieldErrors) {
        this(code, detail, fieldErrors, null);
    }

    public DomainException(ErrorCode code, String detail, List<FieldError> fieldErrors, Throwable cause) {
        super(detail, cause);
        this.code = Objects.requireNonNull(code, "code must not be null");
        this.fieldErrors = List.copyOf(fieldErrors);
    }

    public ErrorCode code() {
        return code;
    }

    public List<FieldError> fieldErrors() {
        return fieldErrors;
    }

    public record FieldError(String field, String message, String current, List<String> allowed) {
        public FieldError {
            Objects.requireNonNull(field, "field must not be null");
            allowed = allowed == null ? List.of() : List.copyOf(allowed);
        }

        public static FieldError of(String field, String message) {
            return new FieldError(field, message, null, List.of());
        }
    }
}
```

Design points worth understanding:

- **It carries a `ErrorCode`, not a category.** The category is reachable as
  `ex.code().category()`. A code is strictly more information than a category, and some
  mappings need the extra precision — `CONFLICT` splits by code, which Part 5 covers.
- **The occurrence message lives in the exception's own `getMessage()`**, passed to
  `super(detail, cause)`. That is the `detail` field of the wire format — specific to *this*
  failure, as opposed to `title()`, which is constant per code. Keeping them separate is an
  explicit requirement of `ERROR_HANDLING.md` §4.
- **`FieldError` is a nested record.** Java records give you an immutable data carrier in
  one line, which is exactly right for a value that is only ever read.
- **`List.copyOf` in the compact constructor** makes the list immutable, so a caller cannot
  mutate the exception's state after construction.

### Naming: why `DomainException` and not `FactoryOsException`

The obvious name would be `FactoryOsException`. We are not using it, and the reason
generalises to every class you name in this codebase.

The package is already `com.factoryos.common.error`. A class named `FactoryOsException`
would spell the project name **twice**, and the second spelling adds no information the
first did not carry. The same mistake in full would be
`com.factoryos.common.error.FactoryOsException` — the redundancy is visible once you write
it out.

More concretely: **the name should say what makes this class special.** `DomainException`
says it belongs to the domain layer and, by extension, knows nothing about transport —
which is the class's defining property. `FactoryOsException` says only "this is ours",
which the package already said.

> **Consistency note.** Every type in this package carries its scope in the name:
> `ErrorCategory`, `ErrorSeverity`, `ErrorCode`, `ErrorKey`, `DomainException`. The pattern
> is *<Scope><Thing>* — so a future reader can tell at a glance that `ErrorSeverity` and
> `DomainException` belong to the same design.

### Verify it compiles

```bash
cd services && mvn -o -B test -pl ../libs/factoryos-common-error -am
```

You should see `BUILD SUCCESS`. The module reports **"No tests to run"** — you have not
written any yet. That is expected, not a failure.

**The `-am` flag explained.** `-pl` (project list) means "build only this module". `-am`
(also-make) means "and also build anything it depends on". This module has no internal
dependencies, so `-am` changes nothing here — but the adapters module in Part 4 *does*
depend on this one, and there `-am` is required. Getting into the habit now saves a
confusing error later.

> **If compilation fails**, check for an accidental import. A stray
> `import org.springframework...` in this module is a design error, not a typo — the
> dependency is deliberately absent, so the build will fail with a missing class rather
> than a warning. That is the constraint working as intended.

### Check yourself

- All six files exist under
  `libs/factoryos-common-error/src/main/java/com/factoryos/common/error/`
- `mvn -o -B test -pl ../libs/factoryos-common-error -am` reports `BUILD SUCCESS`
- Grep your new files for `springframework`, `io.grpc`, `HttpStatus`, and `Status.Code` —
  there should be **zero** matches, and `ErrorCategory` should contain no `int` field at all
- You can explain why `ErrorCode` is an interface and not a class
- You can explain why the category carries no transport code, and where the two projections
  actually live

---


---

[← Part 2 — Maven ground floor](02-maven-ground-floor.md) · [Index](README.md) · [Part 4 — The HTTP adapter →](04-http-adapter.md)
