package com.factoryos.common.error;

import java.util.LinkedHashMap;
import java.util.Map;

import org.springframework.beans.factory.ObjectProvider;
import org.springframework.boot.autoconfigure.AutoConfiguration;
import org.springframework.boot.autoconfigure.condition.ConditionalOnClass;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnWebApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import io.grpc.Status;
import io.grpc.StatusException;

@AutoConfiguration
public class ErrorHandlingAutoConfiguration {

    @Configuration(proxyBeanMethods = false)
    @ConditionalOnClass(StatusException.class)
    static class GrpcAdapters {

        @Bean
        @ConditionalOnMissingBean
        GrpcStatusFactory grpcStatusFactory(ObjectProvider<GrpcStatusOverrideProvider> providers) {
            Map<String, Status.Code> merged = new LinkedHashMap<>();
            providers.orderedStream().forEach(p -> merged.putAll(p.overrides()));
            return new GrpcStatusFactory(merged);
        }

        @Bean
        @ConditionalOnMissingBean
        GrpcErrorAdvice grpcErrorAdvice(GrpcStatusFactory statusFactory) {
            return new GrpcErrorAdvice(statusFactory);
        }
    }

    @Configuration(proxyBeanMethods = false)
    @ConditionalOnClass(name = "org.springframework.web.servlet.DispatcherServlet")
    @ConditionalOnWebApplication
    static class WebAdapters {

        @Bean
        @ConditionalOnMissingBean
        ProblemDetailExceptionHandler problemDetailExceptionHandler() {
            return new ProblemDetailExceptionHandler();
        }

        @Bean
        @ConditionalOnMissingBean
        TraceIdFilter traceIdFilter() {
            return new TraceIdFilter();
        }
    }
}
