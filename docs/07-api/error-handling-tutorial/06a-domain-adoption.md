[Index](README.md) · [← Part 5 — The gRPC adapter](05-grpc-adapter.md)

# Part 6a — Adopting it: the domain side

The libraries exist. Now `production-service` has to use them.

This part changes **only** `production-service` code, and only its error classification. It
does not add dependencies or register adapters — that is Part 6b.

### `WorkOrderErrorCode` — this domain's vocabulary

Create
`services/production-service/src/main/java/com/factoryos/production/WorkOrderErrorCode.java`:

```java
package com.factoryos.production;

import com.factoryos.common.error.ErrorCategory;
import com.factoryos.common.error.ErrorCode;
import com.factoryos.common.error.ErrorSeverity;

/**
 * Error codes owned by the Production bounded context.
 *
 * <p>
 * Domain codes live with their domain, not in the shared library — a shared code
 * must be resolvable by every service, and these are meaningful only here.
 */
public enum WorkOrderErrorCode implements ErrorCode {

    WORK_ORDER_NOT_FOUND(ErrorCategory.NOT_FOUND,
            "Work order not found", ErrorSeverity.INFO, false),

    WORK_ORDER_INVALID_TRANSITION(ErrorCategory.CONFLICT,
            "Invalid state transition", ErrorSeverity.WARNING, false),

    WORK_ORDER_ALREADY_EXISTS(ErrorCategory.CONFLICT,
            "Work order already exists", ErrorSeverity.WARNING, false),

    /**
     * A lost optimistic lock. It shares the {@code CONFLICT} category with the two
     * codes above, but its gRPC status differs: the category default is
     * FAILED_PRECONDITION (9), while a retryable concurrency loss is ABORTED (10).
     * That override is declared in the gRPC adapter's table (see below), not by the
     * code's name. See ERROR_HANDLING.md section 3.2.
     */
    WORK_ORDER_CONCURRENT_MODIFICATION(ErrorCategory.CONFLICT,
            "Work order was modified concurrently", ErrorSeverity.WARNING, true);

    private final ErrorCategory category;
    private final String title;
    private final ErrorSeverity severity;
    private final boolean retryable;

    WorkOrderErrorCode(ErrorCategory category, String title, ErrorSeverity severity, boolean retryable) {
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

**Why this lives in the service, not `libs/`.** `ERROR_HANDLING.md` §6 is explicit: shared
codes go centrally, domain codes per domain, and *"domain codes are owned by the domain that
defines them and must never be reused across domains."* A `WORK_ORDER_NOT_FOUND` in the
quality service would be meaningless. Only codes every service can emit belong in
`CommonErrorCode`.

Note the enum has no `code()` method — the `ErrorCode` interface's `default` supplies it from
the constant name, which is why `WORK_ORDER_CONCURRENT_MODIFICATION` is spelled identically to
the contract.

### Declaring the gRPC override

`WORK_ORDER_CONCURRENT_MODIFICATION` is the one code here whose gRPC status differs from its
category default. The adapter module cannot import this enum, so the service contributes the
override through the provider seam from Part 5. Create
`services/production-service/src/main/java/com/factoryos/production/WorkOrderErrorOverrides.java`:

```java
package com.factoryos.production;

import java.util.Map;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import com.factoryos.common.error.GrpcStatusOverrideProvider;

import io.grpc.Status;

/**
 * A lost optimistic lock is ABORTED (10); every other CONFLICT keeps the default
 * FAILED_PRECONDITION (9). Keyed by code string, so the shared adapter never
 * imports this service's enum. See ERROR_HANDLING.md section 3.2.
 */
@Configuration
public class WorkOrderErrorOverrides {

    @Bean
    public GrpcStatusOverrideProvider workOrderGrpcOverrides() {
        return () -> Map.of(
                WorkOrderErrorCode.WORK_ORDER_CONCURRENT_MODIFICATION.code(),
                Status.Code.ABORTED);
    }
}
```

`WorkOrderErrorCode.WORK_ORDER_CONCURRENT_MODIFICATION.code()` yields the string
`"WORK_ORDER_CONCURRENT_MODIFICATION"`. Using the enum (not a literal) keeps the key from
drifting from the code it names.

### The cursor defect

Now a real bug. `PAGINATION_DESIGN.md` §5.5 requires **three distinct** cursor error codes:

| Check | Required code |
|---|---|
| Missing `v1.` prefix, or wrong version | `UNSUPPORTED_CURSOR_VERSION` |
| Malformed Base64 or JSON | `INVALID_CURSOR` |
| Keys don't match the query's sort keys | `CURSOR_SORT_KEY_MISMATCH` |
| A value can't be parsed as its key type | `INVALID_CURSOR` |

But `Cursor.decode` currently throws **one** exception type carrying only a message. Open it
and count the throw sites:

```bash
grep -n "throw new InvalidCursorException" \
  services/production-service/src/main/java/com/factoryos/production/repository/pagination/Cursor.java
```

Five sites, all indistinguishable to a client — and all five currently surface as **HTTP 500**,
as you saw captured in Part 1. A client cannot tell a version problem from a key mismatch, so
it cannot decide whether to re-fetch a fresh cursor or fix its query.

### Refactor `InvalidCursorException`

Replace the whole file at
`services/production-service/src/main/java/com/factoryos/production/repository/pagination/InvalidCursorException.java`:

```java
package com.factoryos.production.repository.pagination;

import java.util.List;

import com.factoryos.common.error.CommonErrorCode;
import com.factoryos.common.error.DomainException;
import com.factoryos.common.error.ErrorCode;
import com.factoryos.common.error.ErrorKey;

/**
 * Thrown when a cursor token is malformed, expired, or contains mismatched keys.
 *
 * <p>
 * Carries a {@link ErrorCode} so the three distinct cursor failures promised by
 * PAGINATION_DESIGN.md section 5.5 stay distinguishable on the wire. Previously this
 * carried only a message, so a client could not tell them apart.
 */
public class InvalidCursorException extends DomainException {

    private InvalidCursorException(ErrorCode code, String detail, List<FieldError> fieldErrors, Throwable cause) {
        super(code, detail, fieldErrors, cause);
    }

    /** Missing or wrong version prefix. */
    public static InvalidCursorException unsupportedVersion() {
        return new InvalidCursorException(CommonErrorCode.UNSUPPORTED_CURSOR_VERSION,
                "Cursor version not supported; expected a 'v1.' prefix.", List.of(), null);
    }

    /** Malformed base64/JSON, missing payload members, or an unparseable value. */
    public static InvalidCursorException malformed(String detail, Throwable cause) {
        return new InvalidCursorException(CommonErrorCode.INVALID_CURSOR, detail, List.of(), cause);
    }

    /** Well-formed, but issued for a different set of sort keys. */
    public static InvalidCursorException sortKeyMismatch(List<String> expected, List<String> actual) {
        return new InvalidCursorException(CommonErrorCode.CURSOR_SORT_KEY_MISMATCH,
                "Cursor does not match this query's sort keys: expected %s but got %s"
                        .formatted(expected, actual),
                List.of(new FieldError(ErrorKey.CURSOR, "cursor sort keys do not match the query",
                        String.join(",", actual), expected)),
                null);
    }
}
```

Two design points:

**It extends `DomainException`.** That single choice means the REST and gRPC adapters you
built in Parts 4 and 5 handle cursor errors **with no changes at all** — `InvalidCursorException`
*is* a `DomainException`, so both `@ExceptionHandler(DomainException.class)` and
`@GrpcExceptionHandler(DomainException.class)` already match it. This is the payoff of putting
classification in a base class.

**The constructor is private; three static factories are the only way in.** Each factory names
a real failure mode and supplies the matching code, so a caller cannot accidentally construct
a cursor error with the wrong classification. `unsupportedVersion()` takes no message because
the wording should be identical every time.

### Update the five throw sites in `Cursor.decode`

The `decode` method compares an incoming cursor against the query's sort keys. Replace its
throws to use the new factories:

```java
        if (!token.startsWith(VERSION_PREFIX)) {
            throw InvalidCursorException.unsupportedVersion();
        }

        try {
            String base64 = token.substring(VERSION_PREFIX.length());
            byte[] json = Base64.getUrlDecoder().decode(base64);
            Map<String, List<String>> payload = MAPPER.readValue(json, MAP_TYPE);

            List<String> actualKeys = payload.get("keys");
            List<String> actualVals = payload.get("vals");

            if (actualKeys == null || actualVals == null) {
                throw InvalidCursorException.malformed(
                        "Cursor payload is missing 'keys' or 'vals'.", null);
            }
            if (actualKeys.size() != actualVals.size()) {
                throw InvalidCursorException.malformed(
                        "Cursor payload has mismatched 'keys' and 'vals' lengths.", null);
            }

            // Validate keys match expected sort keys
            List<String> expectedKeys = sortKeys.stream().map(SortKey::fieldName).toList();
            if (!actualKeys.equals(expectedKeys)) {
                throw InvalidCursorException.sortKeyMismatch(expectedKeys, actualKeys);
            }

            // Parse each value using the corresponding SortKey
            List<Object> values = new ArrayList<>();
            for (int i = 0; i < sortKeys.size(); i++) {
                values.add(sortKeys.get(i).parseValue(actualVals.get(i)));
            }

            return new Cursor(actualKeys, values);
        } catch (InvalidCursorException e) {
            throw e;
        } catch (Exception e) {
            throw InvalidCursorException.malformed("Invalid cursor token: " + e.getMessage(), e);
        }
    }
```

> **Notice what did *not* change: the `catch (InvalidCursorException e) { throw e; }` block.**
> It still compiles and still does exactly what it did before — rethrow without rewrapping.
> That block is what makes the next section a non-issue.

### What you do *not* need to change

`SortKey.parseValue` throws a raw `IllegalArgumentException` when a cursor value cannot be
parsed as its key type:

```java
// services/production-service/.../pagination/SortKey.java
throw new IllegalArgumentException(...);
```

That looks like an unclassified error, and you might expect it on the conversion list. **It is
not.** `Cursor.decode` wraps its whole body in a `try`, and its final `catch (Exception e)`
converts anything that is not already an `InvalidCursorException` into one:

```java
} catch (Exception e) {
    throw InvalidCursorException.malformed("Invalid cursor token: " + e.getMessage(), e);
}
```

So a bad value is already classified as `INVALID_CURSOR` — correct per `PAGINATION_DESIGN.md`
§5.5, which lists "a value can't be parsed as its key type" as `INVALID_CURSOR`. **Leave
`SortKey` alone.** Changing it would be a no-op at best.

### Classify the page-size failures

Two compact constructors still throw raw `IllegalArgumentException`, which is why `?limit=500`
returns a 500. Both are in the same package:

```bash
services/production-service/src/main/java/com/factoryos/production/repository/pagination/
    OffsetPageRequest.java     # page < 0, or pageSize outside 1..100
    CursorPageRequest.java     # pageSize outside 1..100
```

In each, replace the `IllegalArgumentException` throws with a `DomainException` carrying
`PAGE_SIZE_OUT_OF_RANGE`. For `OffsetPageRequest`, the validation becomes:

```java
    public OffsetPageRequest {
        Objects.requireNonNull(sortCriteriaList, "sortCriteriaList must not be null");
        if (sortCriteriaList.isEmpty()) {
            throw new DomainException(CommonErrorCode.MALFORMED_REQUEST,
                    "sortCriteriaList must not be empty");
        }
        if (page < 0) {
            throw new DomainException(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE,
                    "page must be >= 0, got %d".formatted(page),
                    List.of(new DomainException.FieldError(ErrorKey.PAGE, "must be >= 0",
                            String.valueOf(page), List.of())));
        }
        if (pageSize < 1 || pageSize > MAX_PAGE_SIZE) {
            throw new DomainException(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE,
                    "pageSize must be between 1 and %d, got %d".formatted(MAX_PAGE_SIZE, pageSize),
                    List.of(new DomainException.FieldError(ErrorKey.LIMIT,
                            "must be between 1 and %d".formatted(MAX_PAGE_SIZE),
                            String.valueOf(pageSize),
                            List.of("1..%d".formatted(MAX_PAGE_SIZE)))));
        }
    }
```

And in `CursorPageRequest`:

```java
        if (pageSize < 1 || pageSize > MAX_PAGE_SIZE) {
            throw new DomainException(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE,
                    "pageSize must be between 1 and %d, got %d".formatted(MAX_PAGE_SIZE, pageSize),
                    List.of(new DomainException.FieldError(ErrorKey.LIMIT,
                            "must be between 1 and %d".formatted(MAX_PAGE_SIZE),
                            String.valueOf(pageSize),
                            List.of("1..%d".formatted(MAX_PAGE_SIZE)))));
        }
```

You will need the imports in both files:

```java
import java.util.List;

import com.factoryos.common.error.CommonErrorCode;
import com.factoryos.common.error.DomainException;
import com.factoryos.common.error.ErrorKey;
```

> **Careful: these are record compact constructors.** In a record, the compact constructor
> validates and can throw, but must not assign the fields — assignment happens automatically
> after it returns. The existing code already follows this shape, so replacing only the
> `throw` statements is safe.

### Classify the malformed UUID

One more raw `IllegalArgumentException` escapes, in the gRPC service:

```java
// services/production-service/.../grpc/WorkOrderGrpcService.java
UUID workCenterId = req.getWorkCenterId().isEmpty() ? null : UUID.fromString(req.getWorkCenterId());
```

An empty string is handled, but a non-empty malformed one (`"not-a-uuid"`) throws from
`UUID.fromString`. Wrap it so it classifies as a client error rather than a 500:

```java
        UUID workCenterId = null;
        if (!req.getWorkCenterId().isEmpty()) {
            try {
                workCenterId = UUID.fromString(req.getWorkCenterId());
            } catch (IllegalArgumentException e) {
                throw new DomainException(CommonErrorCode.MALFORMED_REQUEST,
                        e);
            }
        }
        String state = req.getState().isEmpty() ? null : req.getState();
```

### Where this belongs — a deliberate seam

You may have noticed something odd. A pagination cursor has nothing to do with manufacturing.
No work order, no material, no machine — it is a query parameter. Yet
`InvalidCursorException` now carries a *domain* code and extends `DomainException`.

That is worth naming rather than glossing over.

`ERROR_HANDLING.md` classifies a bad cursor as `MALFORMED_REQUEST` (malformed input), and
`PAGINATION_DESIGN.md` §5.5 insists a rejected cursor is **always** a client error. So the
standard treats cursor failures as first-class classified errors — while the code that detects
them lives in `repository.pagination`, an infrastructure concern.

We keep it there, and the reason is practical: **the classification is shared, the detection
is not.** Moving `InvalidCursorException` into a domain package would not make it more
domain-meaningful; it would only move a query-parsing detail somewhere less obvious. What
matters is that it carries the shared **code** vocabulary, so every client sees the same
`INVALID_CURSOR` string it would get from any other service.

So this is a **documented seam**, not an oversight: a persistence concern emitting a shared
classification. If you later introduce a proper application layer, this is a natural thing to
revisit — but it is not a defect today, and it does not block anything.

### Update the existing tests — they will fail, and that is correct

The pagination suite is your regression net, but **it will not stay green by itself.** Ten
existing tests assert the *old* behaviour:

```java
assertThrows(IllegalArgumentException.class,
    () -> OffsetPageRequest.of(0, -5, SORT_BY_ID_ASC));
```

The code now throws `DomainException`. So you get:

```
AssertionFailedError: Unexpected exception type thrown,
  expected: <java.lang.IllegalArgumentException>
  but was:  <com.factoryos.common.error.DomainException>
```

**Do not weaken those tests back to `IllegalArgumentException`, and do not delete them.**
They are asserting the defect. Update them to assert the *new* contract — which makes them
stronger, because they can now check the error code as well as the exception type:

```java
@Test
void of_negativePageSize_throwsPageSizeOutOfRange() {
    DomainException ex = assertThrows(DomainException.class,
        () -> OffsetPageRequest.of(0, -5, SORT_BY_ID_ASC));
    assertEquals(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE, ex.code());
    assertEquals(ErrorCategory.VALIDATION, ex.code().category());
}
```

The affected files and tests:

| File | Tests to update |
|---|---|
| `OffsetPageRequestTest` | `of_negativePageSize`, `of_pageSizeOver100`, `of_negativePage` |
| `CursorPageRequestTest` | `of_negativePageSize`, `of_pageSizeOver100`, `of_pageSize101`, `of_emptySortCriteria` |
| `CursorTest` | `decode_missingVersionPrefix`, `decode_wrongVersionPrefix`, `decode_keysMismatch` |

Two of these deserve more than a mechanical rename:

**`decode_keysMismatch`** previously asserted only that the message contained the word
"mismatch". Now it can assert the exact code — and it should, because this is one of the
three cases `PAGINATION_DESIGN.md` §5.5 says must be distinguishable:

```java
@Test
void decode_keysMismatch_throwsCursorSortKeyMismatch() {
    String token = Cursor.of(List.of(SortKey.id()), List.of(UUID.randomUUID())).encode();
    InvalidCursorException ex = assertThrows(InvalidCursorException.class,
        () -> Cursor.decode(token, List.of(SortKey.ofTimestamp("createdAt"), SortKey.id())));
    assertEquals(CommonErrorCode.CURSOR_SORT_KEY_MISMATCH, ex.code());
    assertEquals(1, ex.fieldErrors().size());
    assertEquals("cursor", ex.fieldErrors().get(0).field());
}
```

**`decode_missingVersionPrefix`** and **`decode_wrongVersionPrefix`** both previously asserted
on message text (`contains("version prefix")`). They now assert `UNSUPPORTED_CURSOR_VERSION`.
Asserting on a code is better than asserting on prose: prose changes freely, a code is the
contract.

Add these tests while you are there — they encode *why* the change matters:

```java
@Test
void badPageSize_isClientError_notServerError() {
    DomainException ex = assertThrows(DomainException.class,
        () -> OffsetPageRequest.of(0, 500, SORT_BY_ID_ASC));
    // The category is the client-error class; the HTTP status itself (400, not a
    // 5xx) is the adapter's projection and is asserted in Part 7.
    assertEquals(ErrorCategory.VALIDATION, ex.code().category(),
        "an out-of-range page size is the caller's mistake, so must not be a 5xx");
    assertFalse(ex.code().retryable(), "retrying the identical request cannot succeed");
}

@Test
void badPageSize_carriesFieldLevelDetail() {
    DomainException ex = assertThrows(DomainException.class,
        () -> OffsetPageRequest.of(0, 500, SORT_BY_ID_ASC));
    assertEquals(1, ex.fieldErrors().size());
    assertEquals("limit", ex.fieldErrors().get(0).field());
    assertEquals("500", ex.fieldErrors().get(0).current());
}
```

### Verify it still builds

```bash
cd services && mvn -o -B clean test -pl production-service
```

You are changing how failures are *classified*, not when they occur, so **no test should be
deleted** — only their expectations updated. Expect the whole `production-service` suite green
(**128 tests** at the time of writing; the exact number drifts as tests are added, so treat it
as "all green", not as a target).

> `mvn -o clean` here is deliberate. After changing a module's dependencies, incremental
> compilation can report confusing errors like `cannot access Sort` — it is compiling against
> a stale classpath. A clean rebuild clears it.

### Check yourself

- `WorkOrderErrorCode.java` exists in `com.factoryos.production`
- All five throws in `Cursor.decode` use the new static factories
- The `catch (InvalidCursorException e) { throw e; }` block is **unchanged**
- `SortKey.java` is **unchanged** — and you can explain why
- `mvn -o -B clean test -pl production-service` is green (128 tests at the time of writing)
- You did not delete or weaken any test

---


---

[← Part 5 — The gRPC adapter](05-grpc-adapter.md) · [Index](README.md) · [Part 6b — Adopting it: the transport side →](06b-transport-adoption.md)
