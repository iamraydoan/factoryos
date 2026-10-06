package com.factoryos.common.error;

import static org.assertj.core.api.Assertions.assertThat;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.header;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.jsonPath;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import java.util.Map;

import org.junit.jupiter.api.Test;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import io.grpc.StatusException;

/**
 * Exercises the three adapter classes that {@code TransportProjectionTest} does
 * not reach — {@link ProblemDetailExceptionHandler}, {@link TraceIdFilter} and
 * {@link GrpcErrorAdvice} — proving the advice and filter actually translate a
 * thrown error, not just that their factories project correctly.
 */
class AdapterWiringTest {

    /** A throwaway endpoint so the advice and filter run inside a real request. */
    @RestController
    static class ThrowingController {
        @GetMapping(value = "/boom-domain", produces = "application/json")
        void domain() {
            throw new DomainException(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE, "pageSize was 500");
        }

        @GetMapping(value = "/boom-unhandled", produces = "application/json")
        void unhandled() {
            throw new NullPointerException("internal detail that must NOT leak to the client");
        }

        @GetMapping(value = "/ok", produces = "application/json")
        void ok() {
            // no-op: the happy path must be untouched by the filter/advice
        }
    }

    private static MockMvc mvc() {
        return MockMvcBuilders.standaloneSetup(new ThrowingController())
                .setControllerAdvice(new ProblemDetailExceptionHandler())
                .addFilters(new TraceIdFilter())
                .build();
    }

    @Test
    void domainException_becomesAProblemDetailWithItsCode() throws Exception {
        // PAGE_SIZE_OUT_OF_RANGE is VALIDATION -> 422, per ERROR_HANDLING.md section 2.
        mvc().perform(get("/boom-domain"))
                .andExpect(status().is(422))
                .andExpect(header().string("Content-Type",
                        org.hamcrest.Matchers.containsString("application/problem+json")))
                .andExpect(jsonPath("$.code").value("PAGE_SIZE_OUT_OF_RANGE"))
                .andExpect(jsonPath("$.status").value(422));
    }

    @Test
    void unhandledException_isReportedAsInternalError_neverEchoed() throws Exception {
        mvc().perform(get("/boom-unhandled"))
                .andExpect(status().isInternalServerError())
                .andExpect(jsonPath("$.code").value("INTERNAL_ERROR"))
                .andExpect(jsonPath("$.detail").value(
                        "An internal error occurred. Quote the traceId when contacting support."));
    }

    @Test
    void traceIdFilter_setsTheResponseHeader_andHonoursAnIncomingOne() throws Exception {
        mvc().perform(get("/ok"))
                .andExpect(status().isOk())
                .andExpect(header().exists(TraceIdFilter.TRACE_ID_HEADER));

        mvc().perform(get("/ok").header(TraceIdFilter.TRACE_ID_HEADER, "caller-supplied-id"))
                .andExpect(header().string(TraceIdFilter.TRACE_ID_HEADER, "caller-supplied-id"));
    }

    @Test
    void traceId_isPresentInTheProblemDetail() throws Exception {
        mvc().perform(get("/boom-unhandled"))
                .andExpect(header().exists(TraceIdFilter.TRACE_ID_HEADER))
                .andExpect(jsonPath("$.traceId").isNotEmpty());
    }

    @Test
    void grpcAdvice_projectsADomainExceptionOntoStatus() {
        GrpcErrorAdvice advice = new GrpcErrorAdvice(new GrpcStatusFactory(Map.of()));

        StatusException mapped = advice.handleDomainException(
                new DomainException(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE, "pageSize was 500"));
        assertThat(mapped.getStatus().getCode()).isEqualTo(io.grpc.Status.Code.INVALID_ARGUMENT);
    }

    @Test
    void grpcAdvice_catchAll_reportsInternal_neverTheRawMessage() {
        GrpcErrorAdvice advice = new GrpcErrorAdvice(new GrpcStatusFactory(Map.of()));

        StatusException mapped = advice.handleUnexpected(new NullPointerException("secret detail"));
        assertThat(mapped.getStatus().getCode()).isEqualTo(io.grpc.Status.Code.INTERNAL);
        assertThat(mapped.getStatus().getDescription()).doesNotContain("secret detail");
    }
}
