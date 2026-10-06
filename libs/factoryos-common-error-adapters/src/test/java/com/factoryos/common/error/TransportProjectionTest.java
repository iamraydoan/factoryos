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

/**
 * Asserts that both adapters project the taxonomy correctly — and, crucially,
 * that they agree with each other.
 */
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

        @Override
        public ErrorCategory category() {
            return category;
        }

        @Override
        public String title() {
            return name();
        }

        @Override
        public ErrorSeverity severity() {
            return severity;
        }

        @Override
        public boolean retryable() {
            return retryable;
        }
    }

    private static DomainException ex(ErrorCode code) {
        return new DomainException(code, "detail for " + code.code());
    }

    @Nested
    class HttpProjection {

        @Test
        void status_comesFromTheCategory_notTheException() {
            // PAGE_SIZE_OUT_OF_RANGE is VALIDATION (the request parsed, the value is
            // out of range) → 422, per ERROR_HANDLING.md section 2. Not a 500, and not
            // the 400 that malformed input (MALFORMED_REQUEST) gets.
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
            List<Map<String, Object>> errors =
                    (List<Map<String, Object>>) problem.getProperties().get("errors");
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
            GrpcStatusFactory bare = new GrpcStatusFactory(Map.of());
            assertThat(bare.toStatusCode(TestCode.WORK_ORDER_CONCURRENT_MODIFICATION))
                    .isEqualTo(io.grpc.Status.Code.FAILED_PRECONDITION);
        }
    }

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
