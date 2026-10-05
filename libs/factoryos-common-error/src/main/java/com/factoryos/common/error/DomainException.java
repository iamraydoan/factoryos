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
