package com.factoryos.production.repository.pagination;

import java.util.List;

import com.factoryos.common.error.CommonErrorCode;
import com.factoryos.common.error.DomainException;
import com.factoryos.common.error.ErrorCode;
import com.factoryos.common.error.ErrorKey;

/**
 * Thrown when a cursor token is malformed, expired, or contains mismatched
 * keys.
 *
 * <p>
 * Carries a {@link ErrorCode} so the three distinct cursor failures promised by
 * PAGINATION_DESIGN.md section 5.5 stay distinguishable on the wire. Previously
 * this carried only a message, so a client could not tell them apart.
 */
public class InvalidCursorException extends DomainException {
    private InvalidCursorException(ErrorCode code, String detail, List<FieldError> fieldErrors, Throwable cause) {
        super(code, detail, fieldErrors, cause);
    }

    public static InvalidCursorException unsupportedVersion() {
        return new InvalidCursorException(CommonErrorCode.UNSUPPORTED_CURSOR_VERSION,
                "Cursor version not supported; expected a 'v1.' prefix.", List.of(), null);
    }

    public static InvalidCursorException malformed(String detail, Throwable cause) {
        return new InvalidCursorException(CommonErrorCode.INVALID_CURSOR, detail, List.of(), cause);
    }

    public static InvalidCursorException sortKeyMismatch(List<String> expected, List<String> actual) {
        return new InvalidCursorException(CommonErrorCode.CURSOR_SORT_KEY_MISMATCH,
                "Cursor does not match this query's sort keys: expected %s but got %s"
                        .formatted(expected, actual),
                List.of(new FieldError(ErrorKey.CURSOR, "cursor sort keys do not match the query",
                        String.join(",", actual), expected)),
                null);
    }
}
