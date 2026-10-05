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
