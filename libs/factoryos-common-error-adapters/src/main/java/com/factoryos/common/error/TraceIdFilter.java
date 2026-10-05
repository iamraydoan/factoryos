package com.factoryos.common.error;

import java.io.IOException;
import java.util.UUID;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;

import org.springframework.web.filter.OncePerRequestFilter;

/**
 * Assigns a correlation id to every request so a client report can be matched
 * to a
 * server log entry. Honours an incoming {@code X-Trace-Id} if present.
 */
public class TraceIdFilter extends OncePerRequestFilter {

    public static final String TRACE_ID_HEADER = "X-Trace-Id";
    private static final ThreadLocal<String> CURRENT = new ThreadLocal<>();

    public static String currentTraceId() {
        return CURRENT.get();
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain chain)
            throws ServletException, IOException {
        String incoming = request.getHeader(TRACE_ID_HEADER);
        String traceId = incoming == null || incoming.isBlank() ? UUID.randomUUID().toString() : incoming;
        CURRENT.set(traceId);
        response.setHeader(TRACE_ID_HEADER, traceId);
        try {
            chain.doFilter(request, response);
        } finally {
            CURRENT.remove();
        }
    }
}
