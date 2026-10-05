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

    @Override
    public ErrorCategory category() {
        return category;
    }

    @Override
    public String title() {
        return title;
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
