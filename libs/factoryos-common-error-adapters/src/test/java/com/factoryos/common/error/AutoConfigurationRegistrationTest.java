package com.factoryos.common.error;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.api.Test;
import org.springframework.boot.autoconfigure.AutoConfigurations;
import org.springframework.boot.test.context.FilteredClassLoader;
import org.springframework.boot.test.context.runner.WebApplicationContextRunner;

import io.grpc.StatusException;

/**
 * Guards the wiring that {@code AdapterWiringTest} cannot see: it builds MockMvc
 * by hand, so a missing {@code META-INF/spring/...AutoConfiguration.imports}
 * leaves it green while the real service returns 500s for every error.
 *
 * <p>
 * This boots an actual context through the auto-configuration entry point, so
 * deleting the imports file — or dropping a bean from the configuration — fails
 * here instead of in production.
 */
class AutoConfigurationRegistrationTest {

    private final WebApplicationContextRunner runner = new WebApplicationContextRunner()
            .withConfiguration(AutoConfigurations.of(ErrorHandlingAutoConfiguration.class));

    @Test
    void registersTheHttpAdapters() {
        runner.run(context -> {
            assertThat(context).hasSingleBean(ProblemDetailExceptionHandler.class);
            assertThat(context).hasSingleBean(TraceIdFilter.class);
            assertThat(context).hasSingleBean(GrpcStatusFactory.class);
        });
    }

    @Test
    void importsFileNamesTheConfiguration_soSpringBootDiscoversIt() throws Exception {
        try (var in = getClass().getClassLoader().getResourceAsStream(
                "META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports")) {
            assertThat(in).as("META-INF/spring/...AutoConfiguration.imports must be on the classpath")
                    .isNotNull();
            String contents = new String(in.readAllBytes(), java.nio.charset.StandardCharsets.UTF_8);
            assertThat(contents.lines().map(String::trim).filter(s -> !s.isEmpty()).toList())
                    .containsExactly(ErrorHandlingAutoConfiguration.class.getName());
        }
    }

    @Test
    void aServiceOverrideWins() {
        runner.withBean(ProblemDetailExceptionHandler.class, () -> new ProblemDetailExceptionHandler())
                .run(context -> assertThat(context).hasSingleBean(ProblemDetailExceptionHandler.class));
    }

    @Test
    void withoutGrpcOnTheClasspath_theRestStillStarts() {
        // The promise of the nested @Configuration: a service with no gRPC layer
        // loses the gRPC beans but still starts and keeps its web adapters.
        runner.withClassLoader(new FilteredClassLoader(StatusException.class))
                .run(context -> {
                    assertThat(context).hasNotFailed();
                    assertThat(context).doesNotHaveBean(GrpcStatusFactory.class);
                    assertThat(context).doesNotHaveBean(GrpcErrorAdvice.class);
                    assertThat(context).hasSingleBean(ProblemDetailExceptionHandler.class);
                });
    }
}
