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
