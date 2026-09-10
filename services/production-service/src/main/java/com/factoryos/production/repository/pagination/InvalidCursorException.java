package com.factoryos.production.repository.pagination;

/**
 * Thrown when a cursor token is malformed, expired, or contains mismatched keys.
 *
 * <p>Unchecked so callers only catch it when they want to — the gRPC layer
 * translates it into {@code INVALID_ARGUMENT} automatically.
 */
public class InvalidCursorException extends RuntimeException {
    public InvalidCursorException(String message) {
        super(message);
    }

    public InvalidCursorException(String message, Throwable cause) {
        super(message, cause);
    }
}
