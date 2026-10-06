package com.factoryos.common.error;

import java.util.List;

import org.springframework.grpc.server.advice.GrpcAdvice;
import org.springframework.grpc.server.advice.GrpcExceptionHandler;

import io.grpc.StatusException;

@GrpcAdvice
public class GrpcErrorAdvice {
    private final GrpcStatusFactory statusFactory;

    public GrpcErrorAdvice(GrpcStatusFactory statusFactory) {
        this.statusFactory = statusFactory;
    }

    @GrpcExceptionHandler(DomainException.class)
    public StatusException handleDomainException(DomainException ex) {
        return statusFactory.toStatusException(ex);
    }

    @GrpcExceptionHandler(Exception.class)
    public StatusException handleUnexpected(Exception ex) {
        DomainException internal = new DomainException(
                CommonErrorCode.INTERNAL_ERROR,
                "An internal error occurred. Quote the traceId when contacting support.",
                List.of(),
                ex);
        return statusFactory.toStatusException(internal);
    }
}
