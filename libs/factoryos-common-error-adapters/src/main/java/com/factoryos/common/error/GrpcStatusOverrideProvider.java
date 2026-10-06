package com.factoryos.common.error;

import java.util.Map;

import io.grpc.Status.Code;

@FunctionalInterface
public interface GrpcStatusOverrideProvider {
    Map<String, Code> overrides();
}
