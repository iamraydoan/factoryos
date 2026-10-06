package com.factoryos.common.error;

import static org.assertj.core.api.Assertions.assertThat;

import java.util.Arrays;
import java.util.stream.Stream;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.EnumSource;

/**
 * Asserts the shared vocabulary — the part that must be identical across every
 * service and language. The transport projections themselves are asserted in the
 * adapters module, where the {@code statusOf}/{@code codeOf} switches live.
 */
class ErrorTaxonomyTest {

    @Test
    void everyCategory_isMapped_sinceTheMappingMustBeExhaustive() {
        // Adding a category means editing both adapters — that is the design.
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
}
