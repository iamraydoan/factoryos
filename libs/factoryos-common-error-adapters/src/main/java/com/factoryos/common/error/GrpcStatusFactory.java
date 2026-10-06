package com.factoryos.common.error;

import java.util.Map;

import com.google.protobuf.Any;
import com.google.rpc.BadRequest;
import com.google.rpc.ErrorInfo;
import com.google.rpc.Status;

import io.grpc.Status.Code;
import io.grpc.StatusException;
import io.grpc.protobuf.StatusProto;

public class GrpcStatusFactory {
    private static final String DOMAIN = "factoryos";
    private final Map<String, Code> overrides;

    public GrpcStatusFactory(Map<String, Code> overrides) {
        this.overrides = Map.copyOf(overrides);
    }

    static Code codeOf(ErrorCategory category) {
        return switch (category) {
            case MALFORMED_REQUEST, VALIDATION -> Code.INVALID_ARGUMENT;
            case AUTHENTICATION -> Code.UNAUTHENTICATED;
            case AUTHORIZATION -> Code.PERMISSION_DENIED;
            case NOT_FOUND -> Code.NOT_FOUND;
            case CONFLICT -> Code.FAILED_PRECONDITION;
            case RATE_LIMIT -> Code.RESOURCE_EXHAUSTED;
            case DEPENDENCY -> Code.UNAVAILABLE;
            case INTERNAL -> Code.INTERNAL;
        };
    }

    public Code toStatusCode(ErrorCode code) {
        Code override = overrides.get(code.code());
        return override != null ? override : codeOf(code.category());
    }

    StatusException toStatusException(DomainException ex) {
        Code grpcCode = toStatusCode(ex.code());
        ErrorInfo.Builder info = ErrorInfo.newBuilder()
                .setReason(ex.code().code())
                .setDomain(DOMAIN)
                .putMetadata("category", ex.code().category().name())
                .putMetadata("retryable", Boolean.toString(ex.code().retryable()));

        Status.Builder status = Status.newBuilder()
                .setCode(grpcCode.value())
                .setMessage(ex.getMessage())
                .addDetails(Any.pack(info.build()));

        if (!ex.fieldErrors().isEmpty()) {
            BadRequest.Builder bad = BadRequest.newBuilder();
            for (DomainException.FieldError fe : ex.fieldErrors()) {
                BadRequest.FieldViolation.Builder v = BadRequest.FieldViolation.newBuilder()
                        .setField(fe.field());
                if (fe.message() != null) {
                    v.setDescription(fe.message());
                }
                bad.addFieldViolations(v);
            }
            status.addDetails(Any.pack(bad.build()));
        }

        return StatusProto.toStatusException(status.build());
    }
}
