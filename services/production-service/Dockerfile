# syntax=docker/dockerfile:1
# FactoryOS production-service — Spring Boot 4 / Java 21.
# Build context MUST be the repository root: the build resolves the parent POM
# (services/pom.xml), sibling libs (libs/factoryos-common-error*) and the
# Protobuf sources (api/contracts/production/v1) from the repo layout.

FROM maven:3.9-eclipse-temurin-21 AS build
WORKDIR /build

# Reactor metadata first for better layer caching. Every module listed in
# services/pom.xml needs at least its pom.xml present or Maven aborts the
# reactor — only production-service (and libs) sources are compiled.
COPY services/pom.xml services/pom.xml
COPY services/production-service/pom.xml services/production-service/pom.xml
COPY services/warehouse-service/pom.xml services/warehouse-service/pom.xml
COPY services/quality-service/pom.xml services/quality-service/pom.xml
COPY services/maintenance-service/pom.xml services/maintenance-service/pom.xml
COPY services/planning-service/pom.xml services/planning-service/pom.xml
COPY libs/factoryos-common-error/pom.xml libs/factoryos-common-error/pom.xml
COPY libs/factoryos-common-error-adapters/pom.xml libs/factoryos-common-error-adapters/pom.xml
# Protobuf sources for stub generation. google/protobuf/*.proto well-known
# imports (e.g. timestamp.proto) resolve from the plugin itself.
COPY api/contracts/production/v1 api/contracts/production/v1

# Pre-fetch dependencies; failure here is non-fatal (offline best-effort).
WORKDIR /build/services
RUN mvn -q -B -pl production-service -am dependency:go-offline || true

WORKDIR /build
COPY libs libs
COPY services/production-service/src services/production-service/src

# -am also builds the libs modules the service depends on.
WORKDIR /build/services
RUN mvn -q -B -pl production-service -am package -DskipTests \
    && cp production-service/target/production-service-*.jar /app.jar

FROM eclipse-temurin:21-jre-jammy AS runtime
RUN useradd -r -u 10001 appuser
USER appuser
COPY --from=build /app.jar /app/app.jar
EXPOSE 3001 4001
ENTRYPOINT ["java", "-jar", "/app/app.jar"]
