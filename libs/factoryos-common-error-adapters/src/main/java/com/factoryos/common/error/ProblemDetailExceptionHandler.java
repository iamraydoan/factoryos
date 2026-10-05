package com.factoryos.common.error;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.slf4j.event.Level;
import org.springframework.http.ProblemDetail;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.context.request.WebRequest;

@RestControllerAdvice
public class ProblemDetailExceptionHandler {

    private static final Logger log = LoggerFactory.getLogger(ProblemDetailExceptionHandler.class);
    private static final String INTERNAL_DETAIL = "An internal error occurred. Quote the traceId when contacting support.";

    @ExceptionHandler(DomainException.class)
    public ResponseEntity<ProblemDetail> handleDomainException(DomainException ex, WebRequest request) {
        ErrorCode code = ex.code();
        String instance = request.getDescription(false).replaceFirst("^uri=", "");
        String traceId = TraceIdFilter.currentTraceId();
        log.atLevel(toSlf4jLevel(code.severity()))
                .setMessage("Domain error {code} on {instance}")
                .addKeyValue("code", code.code())
                .addKeyValue("category", code.category())
                .addKeyValue("traceId", traceId)
                .setCause(needsStackTrace(code.severity()) ? ex : null)
                .log();
        ProblemDetail problem = ProblemDetailFactory.from(ex, instance, traceId);
        return ResponseEntity.status(ProblemDetailFactory.statusOf(code.category())).body(problem);
    }

    @ExceptionHandler(Exception.class)
    public ResponseEntity<ProblemDetail> handleUnexpected(Exception ex, WebRequest request) {
        String instance = request.getDescription(false).replaceFirst("^uri=", "");
        String traceId = TraceIdFilter.currentTraceId();

        log.atError()
                .setMessage("Unhandled exception")
                .addKeyValue("instance", instance)
                .addKeyValue("traceId", traceId)
                .setCause(ex)
                .log();

        ProblemDetail problem = ProblemDetailFactory
                .from(new DomainException(CommonErrorCode.INTERNAL_ERROR, INTERNAL_DETAIL), instance, traceId);
        return ResponseEntity.status(ProblemDetailFactory.statusOf(CommonErrorCode.INTERNAL_ERROR.category()))
                .body(problem);
    }

    private static Level toSlf4jLevel(ErrorSeverity severity) {
        return switch (severity) {
            case INFO -> Level.INFO;
            case WARNING -> Level.WARN;
            case ERROR, CRITICAL -> Level.ERROR;
        };
    }

    private static boolean needsStackTrace(ErrorSeverity severity) {
        return severity == ErrorSeverity.ERROR || severity == ErrorSeverity.CRITICAL;
    }
}
