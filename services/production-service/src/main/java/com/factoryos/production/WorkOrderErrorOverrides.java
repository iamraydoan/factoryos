package com.factoryos.production;

import java.util.Map;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import com.factoryos.common.error.GrpcStatusOverrideProvider;

import io.grpc.Status;

@Configuration
public class WorkOrderErrorOverrides {
    @Bean
    public GrpcStatusOverrideProvider workOrderGrpcOverrides() {
        return () -> Map.of(
                WorkOrderErrorCode.WORK_ORDER_CONCURRENT_MODIFICATION.code(),
                Status.Code.ABORTED);
    }
}
