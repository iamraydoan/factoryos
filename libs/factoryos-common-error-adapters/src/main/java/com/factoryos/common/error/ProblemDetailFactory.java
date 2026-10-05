package com.factoryos.common.error;

import java.net.URI;
import java.time.Instant;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import org.springframework.http.ProblemDetail;

public final class ProblemDetailFactory {

    private ProblemDetailFactory() {
    }

    static int statusOf(ErrorCategory category) {
        return switch (category) {
            case MALFORMED_REQUEST -> 400;
            case VALIDATION -> 422;
            case AUTHENTICATION -> 401;
            case AUTHORIZATION -> 403;
            case NOT_FOUND -> 404;
            case CONFLICT -> 409;
            case RATE_LIMIT -> 429;
            case DEPENDENCY -> 503;
            case INTERNAL -> 500;
        };
    }

    public static ProblemDetail from(DomainException ex, String instancePath, String traceId) {
        ErrorCode code = ex.code();
        ErrorCategory category = code.category();
        ProblemDetail problem = ProblemDetail.forStatus(statusOf(category));
        problem.setType(URI.create(code.typeUri()));
        problem.setTitle(code.title());
        problem.setDetail(ex.getMessage());
        if (instancePath != null) {
            problem.setInstance(URI.create(instancePath));
        }
        problem.setProperty("code", code.code());
        problem.setProperty("retryable", code.retryable());
        if (traceId != null) {
            problem.setProperty("traceId", traceId);
        }
        problem.setProperty("timestamp", Instant.now().toString());
        List<Map<String, Object>> errors = ex.fieldErrors().stream()
                .map(fe -> {
                    Map<String, Object> entry = new LinkedHashMap<>();
                    entry.put("field", fe.field());
                    if (fe.message() != null) {
                        entry.put("message", fe.message());
                    }
                    if (fe.current() != null) {
                        entry.put("current", fe.current());
                    }
                    if (!fe.allowed().isEmpty()) {
                        entry.put("allowed", fe.allowed());
                    }
                    return entry;
                }).toList();
        if (!errors.isEmpty()) {
            problem.setProperty("errors", errors);
        }
        return problem;
    }
}
