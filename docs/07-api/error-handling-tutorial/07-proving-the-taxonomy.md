[Index](README.md) · [← Part 6b — Adopting it: the transport side](06b-transport-adoption.md)

# Part 7 — Proving the taxonomy

Part 6b showed the behaviour by hand. This part makes it permanent, because
`ERROR_HANDLING.md` §6.5 requires it:

> **Assert it.** Every code must have a test asserting its category and both transport
> projections. Codes with a code-level transport override must be asserted explicitly.

Four new test files, split across the two library modules: one in `factoryos-common-error`,
three in `factoryos-common-error-adapters`. (Part 6a separately brought three pre-existing
pagination tests onto `DomainException` — those are the service-side half of the same standard.)

### The taxonomy suite

Create `libs/factoryos-common-error/src/test/java/com/factoryos/common/error/ErrorTaxonomyTest.java`.
This asserts the *vocabulary* — the part that must be identical across every service and
language.

Add JUnit and AssertJ to `libs/factoryos-common-error/pom.xml` if they are not already there:

```xml
        <dependency>
            <groupId>org.junit.jupiter</groupId>
            <artifactId>junit-jupiter</artifactId>
            <scope>test</scope>
        </dependency>
        <dependency>
            <groupId>org.assertj</groupId>
            <artifactId>assertj-core</artifactId>
            <scope>test</scope>
        </dependency>
```

Now the test. The key idea is in the parameterised methods — **iterating over the enum makes
the assertions exhaustive by construction.** A new code cannot be added without being checked.

The full class has ten test methods. Here are the ones that carry the design; write them
first, then add the remainder listed after:

```java
package com.factoryos.common.error;

import static org.assertj.core.api.Assertions.assertThat;

import java.util.Arrays;
import java.util.stream.Stream;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.EnumSource;

class ErrorTaxonomyTest {

    @Test
    void everyCategory_isMapped_sinceTheMappingMustBeExhaustive() {
        // Adding a category means editing both adapters — that is the design.
        // The projections themselves are asserted in the adapters module
        // (TransportProjectionTest), because that is where the switches live.
        assertThat(ErrorCategory.values()).hasSize(9);
    }

    @ParameterizedTest
    @EnumSource(CommonErrorCode.class)
    void everyCommonCode_declaresItsAttributes(CommonErrorCode code) {
        assertThat(code.code()).isNotBlank();
        assertThat(code.category()).isNotNull();
        assertThat(code.title()).isNotBlank();
        assertThat(code.severity()).isNotNull();
    }

    @Test
    void everyCode_nameMatchesItsDeclaredCode() {
        // code() is derived from Enum.name(), so the two can never drift.
        for (CommonErrorCode code : CommonErrorCode.values()) {
            assertThat(code.code()).isEqualTo(code.name());
        }
    }

    @Test
    void sharedCodes_matchTheStandard() {
        assertThat(Arrays.stream(CommonErrorCode.values()).map(Enum::name))
                .containsExactlyInAnyOrder(
                        "INTERNAL_ERROR",
                        "DEPENDENCY_UNAVAILABLE",
                        "RESOURCE_EXHAUSTED",
                        "UNAUTHENTICATED",
                        "PERMISSION_DENIED",
                        "MALFORMED_REQUEST",
                        "MISSING_REQUIRED_FIELD",
                        "INVALID_CURSOR",
                        "UNSUPPORTED_CURSOR_VERSION",
                        "CURSOR_SORT_KEY_MISMATCH",
                        "PAGE_SIZE_OUT_OF_RANGE");
    }

    @ParameterizedTest
    @EnumSource(CommonErrorCode.class)
    void everyCode_derivesAWellFormedTypeUri(CommonErrorCode code) {
        assertThat(code.typeUri())
                .startsWith("https://factoryos.dev/errors/")
                .doesNotContain("_")
                .matches("^https://factoryos\\.dev/errors/[a-z0-9]+(-[a-z0-9]+)*$");
    }

    @Test
    void typeUri_isKebabCaseOfTheCode() {
        assertThat(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE.typeUri())
                .isEqualTo("https://factoryos.dev/errors/page-size-out-of-range");
        assertThat(CommonErrorCode.INTERNAL_ERROR.typeUri())
                .isEqualTo("https://factoryos.dev/errors/internal-error");
    }
}
```

Then add these four, which complete the ten:

```java
    @Test
    void codeNames_followTheNamingConvention() {
        for (CommonErrorCode code : CommonErrorCode.values()) {
            assertThat(code.code())
                    .as("%s must be UPPER_SNAKE_CASE", code)
                    .matches("^[A-Z][A-Z0-9]*(_[A-Z0-9]+)*$");
        }
    }

    @Test
    void retryableFlags_matchTheStandard() {
        assertThat(CommonErrorCode.INTERNAL_ERROR.retryable()).isFalse();
        assertThat(CommonErrorCode.DEPENDENCY_UNAVAILABLE.retryable()).isTrue();
        assertThat(CommonErrorCode.RESOURCE_EXHAUSTED.retryable()).isTrue();
        assertThat(CommonErrorCode.UNAUTHENTICATED.retryable()).isFalse();
        assertThat(CommonErrorCode.PERMISSION_DENIED.retryable()).isFalse();
        assertThat(CommonErrorCode.MALFORMED_REQUEST.retryable()).isFalse();
        assertThat(CommonErrorCode.INVALID_CURSOR.retryable()).isFalse();
        assertThat(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE.retryable()).isFalse();
    }

    @Test
    void severities_matchTheStandard() {
        assertThat(CommonErrorCode.INTERNAL_ERROR.severity()).isEqualTo(ErrorSeverity.CRITICAL);
        assertThat(CommonErrorCode.DEPENDENCY_UNAVAILABLE.severity()).isEqualTo(ErrorSeverity.ERROR);
        assertThat(CommonErrorCode.INVALID_CURSOR.severity()).isEqualTo(ErrorSeverity.INFO);
        assertThat(CommonErrorCode.PERMISSION_DENIED.severity()).isEqualTo(ErrorSeverity.WARNING);
    }

    @Test
    void typeUri_isDerived_neverStored() {
        // Two codes with different names cannot share a URI, because the URI is
        // computed from the code rather than declared alongside it.
        long distinctUris = Stream.of(CommonErrorCode.values())
                .map(CommonErrorCode::typeUri)
                .distinct()
                .count();
        assertThat(distinctUris).isEqualTo(CommonErrorCode.values().length);
    }
```

The `typeUri` assertions are the round-trip check ADR-0007 calls for. The OpenAPI schema is
partly documentation and cannot express everything the runtime produces, so this is what keeps
schema and behaviour from drifting apart.

> The exact category → status mappings are asserted in the **adapters** module
> (`TransportProjectionTest.categoryMappings_matchTheStandard`), because that is where the
> `statusOf`/`codeOf` switches live. This taxonomy suite stays free of transport values —
> it asserts the *vocabulary*, not the projections.

**These parameters multiply into 30 test executions.** Ten methods, of which two are
parameterised: 8 single tests plus (2 × 11 shared codes) = 30.

### The projection suite

Create `libs/factoryos-common-error-adapters/src/test/java/com/factoryos/common/error/TransportProjectionTest.java`.
This asserts that **both** adapters project the taxonomy correctly — and, crucially, that they
agree with each other.

Add JUnit and AssertJ to that module's `pom.xml` the same way:

```xml
        <dependency>
            <groupId>org.junit.jupiter</groupId>
            <artifactId>junit-jupiter</artifactId>
            <scope>test</scope>
        </dependency>
        <dependency>
            <groupId>org.assertj</groupId>
            <artifactId>assertj-core</artifactId>
            <scope>test</scope>
        </dependency>
```

The structure uses `@Nested` classes so failures name the concern. Write these; two more
follow after:

```java
package com.factoryos.common.error;

import static org.assertj.core.api.Assertions.assertThat;

import java.util.List;
import java.util.Map;

import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.springframework.http.ProblemDetail;

import com.google.rpc.ErrorInfo;
import com.google.rpc.Status;

import io.grpc.StatusException;
import io.grpc.protobuf.StatusProto;

class TransportProjectionTest {

    /**
     * The gRPC projector, constructed with the override table a service would
     * register: a lost optimistic lock is ABORTED (10), every other CONFLICT keeps
     * the FAILED_PRECONDITION (9) default. Keyed by the code string, so the shared
     * adapter need not import any service enum.
     */
    private static final GrpcStatusFactory GRPC = new GrpcStatusFactory(Map.of(
            "WORK_ORDER_CONCURRENT_MODIFICATION", io.grpc.Status.Code.ABORTED));

    /** Domain codes mirroring the Production context, for the conflict split. */
    private enum TestCode implements ErrorCode {
        WORK_ORDER_NOT_FOUND(ErrorCategory.NOT_FOUND, ErrorSeverity.INFO, false),
        WORK_ORDER_INVALID_TRANSITION(ErrorCategory.CONFLICT, ErrorSeverity.WARNING, false),
        WORK_ORDER_ALREADY_EXISTS(ErrorCategory.CONFLICT, ErrorSeverity.WARNING, false),
        WORK_ORDER_CONCURRENT_MODIFICATION(ErrorCategory.CONFLICT, ErrorSeverity.WARNING, true);

        private final ErrorCategory category;
        private final ErrorSeverity severity;
        private final boolean retryable;

        TestCode(ErrorCategory category, ErrorSeverity severity, boolean retryable) {
            this.category = category;
            this.severity = severity;
            this.retryable = retryable;
        }

        @Override public ErrorCategory category() { return category; }
        @Override public String title() { return name(); }
        @Override public ErrorSeverity severity() { return severity; }
        @Override public boolean retryable() { return retryable; }
    }

    private static DomainException ex(ErrorCode code) {
        return new DomainException(code, "detail for " + code.code());
    }

    @Nested
    class HttpProjection {

        @Test
        void status_comesFromTheCategory_notTheException() {
            ProblemDetail problem = ProblemDetailFactory.from(
                    ex(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE), "/api/v1/work-orders", "trace-1");
            assertThat(problem.getStatus()).isEqualTo(422);
        }

        @Test
        void carriesEveryRequiredRfc9457Member() {
            ProblemDetail problem = ProblemDetailFactory.from(
                    ex(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE), "/api/v1/work-orders", "trace-1");

            assertThat(problem.getType()).isNotNull();
            assertThat(problem.getTitle()).isEqualTo("Page size out of range");
            assertThat(problem.getStatus()).isEqualTo(422);
            assertThat(problem.getDetail()).isEqualTo("detail for PAGE_SIZE_OUT_OF_RANGE");
            assertThat(problem.getInstance()).hasToString("/api/v1/work-orders");
        }

        @Test
        void type_isTheKebabCaseDerivationOfCode() {
            ProblemDetail problem = ProblemDetailFactory.from(
                    ex(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE), "/x", "trace-1");
            assertThat(problem.getType()).hasToString(
                    "https://factoryos.dev/errors/page-size-out-of-range");
        }

        @Test
        void carriesTheFactoryOsExtensionMembers() {
            ProblemDetail problem = ProblemDetailFactory.from(
                    ex(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE), "/x", "trace-1");
            assertThat(problem.getProperties())
                    .containsEntry("code", "PAGE_SIZE_OUT_OF_RANGE")
                    .containsEntry("retryable", false)
                    .containsEntry("traceId", "trace-1")
                    .containsKey("timestamp");
        }

        @Test
        void titleIsConstantPerCode_whileDetailDescribesThisOccurrence() {
            ProblemDetail a = ProblemDetailFactory.from(
                    new DomainException(CommonErrorCode.INVALID_CURSOR, "first occurrence"), "/a", null);
            ProblemDetail b = ProblemDetailFactory.from(
                    new DomainException(CommonErrorCode.INVALID_CURSOR, "second occurrence"), "/b", null);

            assertThat(a.getTitle()).isEqualTo(b.getTitle());
            assertThat(a.getDetail()).isNotEqualTo(b.getDetail());
        }

        @Test
        void fieldErrors_becomeTheErrorsArray() {
            DomainException withFields = new DomainException(
                    CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE, "too big",
                    List.of(new DomainException.FieldError("limit", "must be between 1 and 100",
                            "500", List.of("1..100"))));

            ProblemDetail problem = ProblemDetailFactory.from(withFields, "/x", null);

            @SuppressWarnings("unchecked")
            List<java.util.Map<String, Object>> errors =
                    (List<java.util.Map<String, Object>>) problem.getProperties().get("errors");
            assertThat(errors).hasSize(1);
            assertThat(errors.get(0))
                    .containsEntry("field", "limit")
                    .containsEntry("message", "must be between 1 and 100")
                    .containsEntry("current", "500")
                    .containsEntry("allowed", List.of("1..100"));
        }

        @Test
        void noErrorsArray_whenThereAreNoFieldErrors() {
            ProblemDetail problem = ProblemDetailFactory.from(
                    ex(CommonErrorCode.INTERNAL_ERROR), "/x", null);
            assertThat(problem.getProperties()).doesNotContainKey("errors");
        }

        @Test
        void everySharedCode_projectsToItsCategoryStatus() {
            for (CommonErrorCode code : CommonErrorCode.values()) {
                ProblemDetail problem = ProblemDetailFactory.from(ex(code), "/x", null);
                // The expected value comes from the adapter's own switch, not from a
                // field on the category — the category carries no status (Part 3).
                assertThat(problem.getStatus())
                        .as("%s", code)
                        .isEqualTo(ProblemDetailFactory.statusOf(code.category()));
            }
        }

        @Test
        void categoryMappings_matchTheStandard() {
            // The one test that pins the exact numbers, including the special cases.
            assertThat(ProblemDetailFactory.statusOf(ErrorCategory.MALFORMED_REQUEST)).isEqualTo(400);
            assertThat(ProblemDetailFactory.statusOf(ErrorCategory.VALIDATION)).isEqualTo(422);
            assertThat(ProblemDetailFactory.statusOf(ErrorCategory.AUTHENTICATION)).isEqualTo(401);
            assertThat(ProblemDetailFactory.statusOf(ErrorCategory.AUTHORIZATION)).isEqualTo(403);
            assertThat(ProblemDetailFactory.statusOf(ErrorCategory.NOT_FOUND)).isEqualTo(404);
            assertThat(ProblemDetailFactory.statusOf(ErrorCategory.CONFLICT)).isEqualTo(409);
            assertThat(ProblemDetailFactory.statusOf(ErrorCategory.RATE_LIMIT)).isEqualTo(429);
            assertThat(ProblemDetailFactory.statusOf(ErrorCategory.DEPENDENCY)).isEqualTo(503);
            assertThat(ProblemDetailFactory.statusOf(ErrorCategory.INTERNAL)).isEqualTo(500);
        }
    }

    @Nested
    class GrpcProjection {

        @Test
        void status_comesFromTheCategory() {
            StatusException mapped = GRPC.toStatusException(
                    ex(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE));
            assertThat(mapped.getStatus().getCode().name()).isEqualTo("INVALID_ARGUMENT");
            assertThat(mapped.getStatus().getCode().value()).isEqualTo(3);
        }

        @Test
        void everySharedCode_projectsToItsCategoryGrpcCode() {
            for (CommonErrorCode code : CommonErrorCode.values()) {
                StatusException mapped = GRPC.toStatusException(ex(code));
                // Same rule as the HTTP side: the expected value is the adapter's
                // projection, not a field on the category.
                assertThat(mapped.getStatus().getCode())
                        .as("%s", code)
                        .isEqualTo(GrpcStatusFactory.codeOf(code.category()));
            }
        }

        @Test
        void categoryMappings_matchTheStandard() {
            assertThat(GrpcStatusFactory.codeOf(ErrorCategory.MALFORMED_REQUEST))
                    .isEqualTo(io.grpc.Status.Code.INVALID_ARGUMENT);
            assertThat(GrpcStatusFactory.codeOf(ErrorCategory.VALIDATION))
                    .isEqualTo(io.grpc.Status.Code.INVALID_ARGUMENT);
            assertThat(GrpcStatusFactory.codeOf(ErrorCategory.AUTHENTICATION))
                    .isEqualTo(io.grpc.Status.Code.UNAUTHENTICATED);
            assertThat(GrpcStatusFactory.codeOf(ErrorCategory.AUTHORIZATION))
                    .isEqualTo(io.grpc.Status.Code.PERMISSION_DENIED);
            assertThat(GrpcStatusFactory.codeOf(ErrorCategory.NOT_FOUND))
                    .as("a missing entity is not a dependency failure")
                    .isEqualTo(io.grpc.Status.Code.NOT_FOUND);
            assertThat(GrpcStatusFactory.codeOf(ErrorCategory.CONFLICT))
                    .isEqualTo(io.grpc.Status.Code.FAILED_PRECONDITION);
            assertThat(GrpcStatusFactory.codeOf(ErrorCategory.RATE_LIMIT))
                    .isEqualTo(io.grpc.Status.Code.RESOURCE_EXHAUSTED);
            assertThat(GrpcStatusFactory.codeOf(ErrorCategory.DEPENDENCY))
                    .isEqualTo(io.grpc.Status.Code.UNAVAILABLE);
            assertThat(GrpcStatusFactory.codeOf(ErrorCategory.INTERNAL))
                    .isEqualTo(io.grpc.Status.Code.INTERNAL);
        }

        @Test
        void notFound_mapsToNotFound_neverUnavailable() {
            StatusException mapped = GRPC.toStatusException(
                    ex(TestCode.WORK_ORDER_NOT_FOUND));
            assertThat(mapped.getStatus().getCode().name())
                    .as("a missing entity is permanent; NOT_FOUND must not invite a retry")
                    .isEqualTo("NOT_FOUND");
            assertThat(mapped.getStatus().getCode().value()).isEqualTo(5).isNotEqualTo(14);
        }

        @Test
        void errorInfo_reasonHoldsTheSameCodeStringRestUses() {
            ErrorInfo info = unpackErrorInfo(GRPC.toStatusException(
                    ex(TestCode.WORK_ORDER_INVALID_TRANSITION)));
            assertThat(info.getReason()).isEqualTo("WORK_ORDER_INVALID_TRANSITION");
            assertThat(info.getDomain()).isEqualTo("factoryos");
        }

        @Test
        void errorInfo_carriesCategoryAndRetryableMetadata() {
            ErrorInfo info = unpackErrorInfo(GRPC.toStatusException(
                    ex(TestCode.WORK_ORDER_INVALID_TRANSITION)));
            assertThat(info.getMetadataMap())
                    .containsEntry("category", "CONFLICT")
                    .containsEntry("retryable", "false");
        }

        @Test
        void detailSurvivesTheTrailer_whichABareStatusWouldLose() {
            // This is the trap from Part 5: Status.X.asException() gives the code, no detail.
            StatusException bare = io.grpc.Status.INVALID_ARGUMENT.asException();
            assertThat(StatusProto.fromThrowable(bare).getDetailsList()).isEmpty();

            StatusException mapped = GRPC.toStatusException(
                    ex(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE));
            assertThat(StatusProto.fromThrowable(mapped).getDetailsList()).isNotEmpty();
        }
    }

    @Nested
    class ConflictSplit {

        @Test
        void concurrentModification_isAborted_andRetryable() {
            StatusException mapped = GRPC.toStatusException(
                    ex(TestCode.WORK_ORDER_CONCURRENT_MODIFICATION));
            assertThat(mapped.getStatus().getCode().name())
                    .as("a lost optimistic lock should be retried as a whole transaction")
                    .isEqualTo("ABORTED");
            assertThat(mapped.getStatus().getCode().value()).isEqualTo(10);
            assertThat(TestCode.WORK_ORDER_CONCURRENT_MODIFICATION.retryable()).isTrue();
        }

        @Test
        void invalidTransition_isFailedPrecondition_andNotRetryable() {
            StatusException mapped = GRPC.toStatusException(
                    ex(TestCode.WORK_ORDER_INVALID_TRANSITION));
            assertThat(mapped.getStatus().getCode().name())
                    .as("retrying a permanent state rejection can never succeed")
                    .isEqualTo("FAILED_PRECONDITION");
            assertThat(mapped.getStatus().getCode().value()).isEqualTo(9);
            assertThat(TestCode.WORK_ORDER_INVALID_TRANSITION.retryable()).isFalse();
        }

        @Test
        void alreadyExists_isFailedPrecondition_notAborted() {
            StatusException mapped = GRPC.toStatusException(
                    ex(TestCode.WORK_ORDER_ALREADY_EXISTS));
            assertThat(mapped.getStatus().getCode().name()).isEqualTo("FAILED_PRECONDITION");
        }

        @Test
        void sameCategory_differentGrpcCodes_whichIsWhyTheCodeDrivesTheMapping() {
            assertThat(TestCode.WORK_ORDER_CONCURRENT_MODIFICATION.category())
                    .isEqualTo(TestCode.WORK_ORDER_INVALID_TRANSITION.category());
            assertThat(GRPC.toStatusCode(TestCode.WORK_ORDER_CONCURRENT_MODIFICATION))
                    .isNotEqualTo(GRPC.toStatusCode(TestCode.WORK_ORDER_INVALID_TRANSITION));
        }

        @Test
        void concurrentModificationOverride_selectsAborted() {
            // Asserted explicitly against the override table, not by name: a rename
            // changes nothing because the override is keyed by the code string, and a
            // *missing* entry merely falls back to the FAILED_PRECONDITION default.
            // See ERROR_HANDLING.md section 3.2.
            assertThat(GRPC.toStatusCode(TestCode.WORK_ORDER_CONCURRENT_MODIFICATION))
                    .isEqualTo(io.grpc.Status.Code.ABORTED);
        }

        @Test
        void aConflictWithNoOverride_fallsBackToTheCategoryDefault() {
            // The negative half: a CONFLICT code absent from the override table gets
            // the category default, never ABORTED. This is the fail-safe property —
            // an unregistered code can't invite a retry storm.
            assertThat(GRPC.toStatusCode(TestCode.WORK_ORDER_INVALID_TRANSITION))
                    .isEqualTo(io.grpc.Status.Code.FAILED_PRECONDITION);
        }

        @Test
        void aFactoryWithNoOverrides_projectsEveryConflictToTheDefault() {
            // The wiring degrades safely: no provider beans → empty table → category
            // defaults. ERROR_HANDLING.md section 3.2.
            GrpcStatusFactory bare = new GrpcStatusFactory(java.util.Map.of());
            assertThat(bare.toStatusCode(TestCode.WORK_ORDER_CONCURRENT_MODIFICATION))
                    .isEqualTo(io.grpc.Status.Code.FAILED_PRECONDITION);
        }
    }

    private static ErrorInfo unpackErrorInfo(StatusException ex) {
        Status status = StatusProto.fromThrowable(ex);
        return status.getDetailsList().stream()
                .filter(any -> any.is(ErrorInfo.class))
                .map(any -> {
                    try {
                        return any.unpack(ErrorInfo.class);
                    } catch (com.google.protobuf.InvalidProtocolBufferException e) {
                        throw new IllegalStateException("ErrorInfo detail could not be decoded", e);
                    }
                })
                .findFirst()
                .orElseThrow(() -> new AssertionError("no ErrorInfo detail on the status"));
    }
}
```

Finally, add the cross-protocol class — the one that proves the design pays off:

```java
    @Nested
    class CrossProtocol {

        @Test
        void theCodeStringIsIdenticalOverBothProtocols() {
            for (TestCode code : TestCode.values()) {
                ProblemDetail problem = ProblemDetailFactory.from(ex(code), "/x", null);
                ErrorInfo info = unpackErrorInfo(GRPC.toStatusException(ex(code)));

                assertThat(info.getReason())
                        .as("%s must be spelled identically on both transports", code)
                        .isEqualTo(problem.getProperties().get("code"));
            }
        }

        @Test
        void retryableIsIdenticalOverBothProtocols() {
            for (TestCode code : TestCode.values()) {
                ProblemDetail problem = ProblemDetailFactory.from(ex(code), "/x", null);
                ErrorInfo info = unpackErrorInfo(GRPC.toStatusException(ex(code)));

                assertThat(info.getMetadataMap().get("retryable"))
                        .isEqualTo(String.valueOf(problem.getProperties().get("retryable")));
            }
        }
    }
```

That is **25 test methods**, and 25 executions — none are parameterised, so the count is
one-to-one. The two `categoryMappings_matchTheStandard` methods (one per projection) are the
tests that pin the exact numbers now that the category carries none.

### The adapter-wiring suite

`TransportProjectionTest` proves the *factories* project correctly — but it calls them directly,
so it never proves the advice and filter actually run inside a request. `AdapterWiringTest`
covers that: it builds MockMvc by hand and drives a throwaway controller through the real
`ProblemDetailExceptionHandler` and `TraceIdFilter`, so a `DomainException` thrown in a request
becomes a Problem Detail and an unhandled exception never echoes its message.

Create `libs/factoryos-common-error-adapters/src/test/java/com/factoryos/common/error/AdapterWiringTest.java`:

```java
    @RestController
    static class ThrowingController {
        @GetMapping(value = "/boom-domain", produces = "application/json")
        void domain() {
            throw new DomainException(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE, "pageSize was 500");
        }

        @GetMapping(value = "/boom-unhandled", produces = "application/json")
        void unhandled() {
            throw new NullPointerException("internal detail that must NOT leak to the client");
        }
    }

    private static MockMvc mvc() {
        return MockMvcBuilders.standaloneSetup(new ThrowingController())
                .setControllerAdvice(new ProblemDetailExceptionHandler())
                .addFilters(new TraceIdFilter())
                .build();
    }

    @Test
    void unhandledException_isReportedAsInternalError_neverEchoed() throws Exception {
        mvc().perform(get("/boom-unhandled"))
                .andExpect(status().isInternalServerError())
                .andExpect(jsonPath("$.code").value("INTERNAL_ERROR"))
                .andExpect(jsonPath("$.detail").value(
                        "An internal error occurred. Quote the traceId when contacting support."));
    }
```

Then the registration suite.

### The registration suite

`AdapterWiringTest` builds its context by hand, so it is green whether or not Spring Boot ever
*discovers* the auto-configuration — delete the `.imports` file and it still passes while a real
service returns 500s. `AutoConfigurationRegistrationTest` closes that hole: it boots a context
**through the auto-configuration entry point** and asserts the imports file names the class.

Create `libs/factoryos-common-error-adapters/src/test/java/com/factoryos/common/error/AutoConfigurationRegistrationTest.java`:

```java
class AutoConfigurationRegistrationTest {

    private final WebApplicationContextRunner runner = new WebApplicationContextRunner()
            .withConfiguration(AutoConfigurations.of(ErrorHandlingAutoConfiguration.class));

    @Test
    void registersTheHttpAdapters() {
        runner.run(context -> {
            assertThat(context).hasSingleBean(ProblemDetailExceptionHandler.class);
            assertThat(context).hasSingleBean(TraceIdFilter.class);
            assertThat(context).hasSingleBean(GrpcStatusFactory.class);
        });
    }

    @Test
    void importsFileNamesTheConfiguration_soSpringBootDiscoversIt() throws Exception {
        try (var in = getClass().getClassLoader().getResourceAsStream(
                "META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports")) {
            assertThat(in).as("META-INF/spring/...AutoConfiguration.imports must be on the classpath")
                    .isNotNull();
            String contents = new String(in.readAllBytes(), java.nio.charset.StandardCharsets.UTF_8);
            assertThat(contents.lines().map(String::trim).filter(s -> !s.isEmpty()).toList())
                    .containsExactly(ErrorHandlingAutoConfiguration.class.getName());
        }
    }

    @Test
    void aServiceOverrideWins() {
        runner.withBean(ProblemDetailExceptionHandler.class, () -> new ProblemDetailExceptionHandler())
                .run(context -> assertThat(context).hasSingleBean(ProblemDetailExceptionHandler.class));
    }

    @Test
    void withoutGrpcOnTheClasspath_theRestStillStarts() {
        // The promise of the nested @Configuration: a service with no gRPC layer
        // loses the gRPC beans but still starts and keeps its web adapters.
        runner.withClassLoader(new FilteredClassLoader(StatusException.class))
                .run(context -> {
                    assertThat(context).hasNotFailed();
                    assertThat(context).doesNotHaveBean(GrpcStatusFactory.class);
                    assertThat(context).doesNotHaveBean(GrpcErrorAdvice.class);
                    assertThat(context).hasSingleBean(ProblemDetailExceptionHandler.class);
                });
    }
}
```

**Why `ConflictSplit` deserves its own nested class.** It is the one place where the mapping
is not derivable from the category alone, so it is the one place a regression would be silent.
`sameCategory_differentGrpcCodes` asserts the *precondition* for the whole design — that two
codes in one category really do project differently — and `concurrentModificationOverride`
pins the exact mechanism: the code's entry in the override table. Its two companions assert the
negative (a conflict *without* an override stays on the default), which is what makes the
mechanism fail-safe rather than merely correct.

**Why `CrossProtocol` exists.** `theCodeStringIsIdenticalOverBothProtocols` is the assertion
that proves the entire design pays off. If the REST `code` and the gRPC `ErrorInfo.reason` ever
diverge, this fails — and a client that switches on `code` stops working on one of the two
protocols.

### Run everything

```bash
cd services && mvn -o -B clean test
```

Expect `BUILD SUCCESS`, with:

| Module | Tests |
|---|---|
| `factoryos-common-error` | 30 |
| `factoryos-common-error-adapters` | 35 |
| `production-service` | 128 |
| the other four Java services | 0 — no test tree yet |

**193 tests, 0 failures.**

The adapters module's 35 break down as 25 in `TransportProjectionTest`, 6 in
`AdapterWiringTest` and 4 in `AutoConfigurationRegistrationTest`.

### Check yourself

- All four new library test files exist, and both modules declare JUnit and AssertJ
- `mvn -o -B clean test` reports **193 tests, 0 failures**
- You can explain why `ConflictSplit` asserts by code and not by category
- You can explain what breaks if `theCodeStringIsIdenticalOverBothProtocols` fails
- You can explain why `AutoConfigurationRegistrationTest` exists even though `AdapterWiringTest`
  already passes

---


---

[← Part 6b — Adopting it: the transport side](06b-transport-adoption.md) · [Index](README.md) · [Part 8 — Rolling it out →](08-rolling-it-out.md)
