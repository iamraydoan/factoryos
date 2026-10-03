# ==============================================================================
# FactoryOS Monorepo Makefile (Optional Developer Convenience Shortcuts)
# ==============================================================================

SHELL := /bin/bash
BIN_DIR := bin

# Explicit path: npm ships an `openapi` alias that collides, so never rely on bare `openapi` on PATH.
OPENAPI_OVERLAY_BIN ?= $(shell go env GOPATH)/bin/openapi

# Compose file sets. Base = infra only (never starts app/docs).
COMPOSE_BASE := -f docker-compose.yml
COMPOSE_SERVICES := -f docker-compose.yml -f docker-compose.services.yml
COMPOSE_DOCS := -f docker-compose.yml -f docker-compose.docs.yml

.PHONY: all help build build-all build-analytics build-ingestion build-edge build-simulator build-resource \
        test test-all test-analytics test-ingestion test-edge test-sdk test-resource \
        test-coverage test-coverage-analytics test-coverage-resource test-coverage-ingestion test-coverage-edge test-coverage-sdk \
        run-analytics run-ingestion run-edge run-simulator run-resource run-production run-production-local \
        setup-production-host setup-production-docker \
        install-tools \
        proto-lint proto-gen openapi-lint openapi-bundle openapi-gateway openapi-gen \
        setup-env-host setup-env-docker setup-env-devcontainer \
        infra-up infra-down infra-ps infra-logs services-up services-down services-logs docs-up docs-down docs-logs clean

all: help

## help: Display available commands
help:
	@echo "FactoryOS Monorepo - Available commands (Optional Convenience):"
	@sed -n "s/^##//p" $(MAKEFILE_LIST) | column -t -s ":" | sed -e "s/^/ /"

## install-tools: Install protobuf, Buf, OpenAPI generator, and Redocly CLI tools
install-tools:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1
	@npm install --global --no-fund --no-audit @redocly/cli@1.34.0
	@go install github.com/speakeasy-api/openapi/cmd/openapi@latest

# ==============================================================================
# Build Targets
# ==============================================================================

## build: Build all Go binaries into $(BIN_DIR)/
build: build-all

## build-all: Build all services, edge runtime, and simulators
build-all: build-analytics build-ingestion build-edge build-simulator build-resource

## build-analytics: Build binary for Analytics Engine into $(BIN_DIR)/analytics-engine
build-analytics:
	@mkdir -p $(BIN_DIR)
	@echo "[BUILD] Compiling services/analytics-engine..."
	@go build -o $(BIN_DIR)/analytics-engine ./services/analytics-engine
	@echo "[BUILD] Success -> $(BIN_DIR)/analytics-engine"

## build-ingestion: Build binary for Ingestion Service into $(BIN_DIR)/ingestion-service
build-ingestion:
	@mkdir -p $(BIN_DIR)
	@echo "[BUILD] Compiling services/ingestion-service..."
	@go build -o $(BIN_DIR)/ingestion-service ./services/ingestion-service
	@echo "[BUILD] Success -> $(BIN_DIR)/ingestion-service"

## build-edge: Build binary for Edge Runtime into $(BIN_DIR)/edge-runtime
build-edge:
	@mkdir -p $(BIN_DIR)
	@echo "[BUILD] Compiling platform/edge-runtime..."
	@go build -o $(BIN_DIR)/edge-runtime ./platform/edge-runtime
	@echo "[BUILD] Success -> $(BIN_DIR)/edge-runtime"

## build-simulator: Build binary for Mock PLC Simulator into $(BIN_DIR)/mock-plc-simulator
build-simulator:
	@mkdir -p $(BIN_DIR)
	@echo "[BUILD] Compiling examples/mock-plc-simulator..."
	@go build -o $(BIN_DIR)/mock-plc-simulator ./examples/mock-plc-simulator
	@echo "[BUILD] Success -> $(BIN_DIR)/mock-plc-simulator"

## build-resource: Build binary for Resource Service into $(BIN_DIR)/resource-service
build-resource:
	@mkdir -p $(BIN_DIR)
	@echo "[BUILD] Compiling services/resource-service..."
	@go build -o $(BIN_DIR)/resource-service ./services/resource-service
	@echo "[BUILD] Success -> $(BIN_DIR)/resource-service"

# ==============================================================================
# Test Targets
# ==============================================================================

## test: Run unit tests across all Go modules
test: test-all

## test-all: Run all unit tests for Analytics Engine, Ingestion Service, Edge Runtime, Platform SDK, and Resource Service
test-all: test-edge test-sdk test-ingestion test-analytics test-resource

## test-analytics: Run unit tests for Analytics Engine (with coverage & race detector)
test-analytics:
	@echo "[TEST] Running tests for analytics-engine..."
	@cd services/analytics-engine && go test -race -cover -v -timeout 60s ./processor/... ./db/... ./consumer/... ./config/...

## test-ingestion: Run unit tests for Ingestion Service
test-ingestion:
	@echo "[TEST] Running tests for ingestion-service..."
	@cd services/ingestion-service && go test -race -cover -v ./...

## test-edge: Run unit tests for Edge Runtime
test-edge:
	@echo "[TEST] Running tests for edge-runtime..."
	@cd platform/edge-runtime && go test -v -cover ./...

## test-sdk: Run unit tests for Platform SDK
test-sdk:
	@echo "[TEST] Running tests for platform-sdk..."
	@cd platform/platform-sdk && go test -v ./...

## test-resource: Run unit tests for Resource Service (with coverage & race detector)
test-resource:
	@echo "[TEST] Running tests for resource-service..."
	@cd services/resource-service && go test -race -cover -v -timeout 60s ./...

## test-coverage: Run tests with coverage for all modules
test-coverage: test-coverage-analytics test-coverage-resource test-coverage-ingestion test-coverage-edge test-coverage-sdk

## test-coverage-analytics: Generate coverage for analytics-engine
test-coverage-analytics:
	@echo "[COVERAGE] analytics-engine..."
	@cd services/analytics-engine && go test -race -coverprofile=coverage.out ./processor/... ./db/... ./consumer/... ./config/... && go tool cover -func=coverage.out

## test-coverage-resource: Generate coverage for resource-service
test-coverage-resource:
	@echo "[COVERAGE] resource-service..."
	@cd services/resource-service && go test -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

## test-coverage-ingestion: Generate coverage for ingestion-service
test-coverage-ingestion:
	@echo "[COVERAGE] ingestion-service..."
	@cd services/ingestion-service && go test -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

## test-coverage-edge: Generate coverage for edge-runtime
test-coverage-edge:
	@echo "[COVERAGE] edge-runtime..."
	@cd platform/edge-runtime && go test -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

## test-coverage-sdk: Generate coverage for platform-sdk
test-coverage-sdk:
	@echo "[COVERAGE] platform-sdk..."
	@cd platform/platform-sdk && go test -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

# ==============================================================================
# Run Targets
# ==============================================================================

## run-analytics: Build and execute Analytics Engine locally
run-analytics: build-analytics
	@echo "[RUN] Starting $(BIN_DIR)/analytics-engine..."
	@$(BIN_DIR)/analytics-engine

## run-ingestion: Build and execute Ingestion Service locally
run-ingestion: build-ingestion
	@echo "[RUN] Starting $(BIN_DIR)/ingestion-service..."
	@$(BIN_DIR)/ingestion-service

## run-edge: Build and execute Edge Runtime locally
run-edge: build-edge
	@echo "[RUN] Starting $(BIN_DIR)/edge-runtime..."
	@$(BIN_DIR)/edge-runtime

## run-simulator: Execute Mock PLC Simulator locally
run-simulator:
	@echo "[RUN] Starting Mock PLC Simulator..."
	@cd examples/mock-plc-simulator && go run main.go

## run-resource: Build and execute Resource Service locally
run-resource: build-resource
	@echo "[RUN] Starting $(BIN_DIR)/resource-service..."
	@$(BIN_DIR)/resource-service

# ==============================================================================
# Java Services (Maven + .env)
# ==============================================================================
# Spring reads OS env, not `.env` files — each target sources
# services/production-service/.env first so DB_HOST/DB_USER/DB_PASSWORD
# reach application.yml. Pick the template matching where the JVM runs:
# .env.example (host, DB at localhost) vs .env.docker.example
# (devcontainer / Docker network, DB at factoryos-db).

## setup-production-host: Copy host-local .env template for production-service (DB at localhost)
setup-production-host:
	@cp services/production-service/.env.example services/production-service/.env
	@echo "[SETUP] production-service .env -> host-local (DB_HOST=localhost)"

## setup-production-docker: Copy docker-network .env template for production-service (DB at factoryos-db)
setup-production-docker:
	@cp services/production-service/.env.docker.example services/production-service/.env
	@echo "[SETUP] production-service .env -> docker network (DB_HOST=factoryos-db)"

# ==============================================================================
# Environment profiles (root .env + services/production-service/.env)
# ==============================================================================
# Three supported profiles — pick the one matching where processes run:
#
#   host         DB=localhost            gateway=host.docker.internal  (service on host)
#   docker       DB=factoryos-db         gateway=production-service    (service as container)
#   devcontainer DB=factoryos-db         gateway=host.docker.internal  (service on host side,
#                                                                      devcontainer joins factoryos_net)

## setup-env-host: Install HOST profile env (root .env + service .env: DB=localhost, gateway=host.docker.internal)
setup-env-host:
	@cp .env.example .env
	@cp services/production-service/.env.example services/production-service/.env
	@echo "[SETUP] HOST profile -> .env + services/production-service/.env (DB=localhost, gateway=host.docker.internal)"

## setup-env-docker: Install DOCKER profile env (root .env + service .env: DB + gateway = container names)
setup-env-docker:
	@cp .env.docker.example .env
	@cp services/production-service/.env.docker.example services/production-service/.env
	@echo "[SETUP] DOCKER profile -> .env + services/production-service/.env (DB + gateway = container names)"

## setup-env-devcontainer: Install DEVCONTAINER profile env (root .env + service .env: DB=name, gateway=host.docker.internal)
setup-env-devcontainer:
	@cp .env.devcontainer.example .env
	@cp services/production-service/.env.devcontainer.example services/production-service/.env
	@echo "[SETUP] DEVCONTAINER profile -> .env + services/production-service/.env (DB=factoryos-db, gateway=host.docker.internal)"

## run-production: Run production-service with .env sourced (works on host and in devcontainer)
run-production:
	@echo "[RUN] Starting production-service (sourcing .env)..."
	@set -a; [ -f services/production-service/.env ] && . services/production-service/.env; set +a; \
	cd services && mvn -pl production-service spring-boot:run

## run-production-local: Run production-service with .env sourced + local profile (show-sql)
run-production-local:
	@echo "[RUN] Starting production-service with local profile (sourcing .env)..."
	@set -a; [ -f services/production-service/.env ] && . services/production-service/.env; set +a; \
	cd services && mvn -pl production-service spring-boot:run -Dspring-boot.run.profiles=local

# ==============================================================================
# Protobuf / Schema Targets
# ==============================================================================

## proto-lint: Lint Protobuf contracts with Buf
proto-lint:
	@echo "[BUF] Linting Protobuf schemas in api/contracts..."
	@cd api/contracts && buf lint

## proto-gen: Generate Go & Java stubs from Protobuf contracts
proto-gen:
	@echo "[BUF] Generating code from api/contracts..."
	@cd api/contracts && buf generate

## openapi-lint: Lint the unified platform OpenAPI contract with Redocly
##   Single entrypoint: api/contracts/openapi/openapi.yaml
openapi-lint:
	@echo "[OPENAPI] Linting api/contracts/openapi/openapi.yaml..."
	@redocly lint api/contracts/openapi/openapi.yaml

## openapi-bundle: Lint and bundle the unified platform contract into dist/ (via Redocly)
##   Input:  api/contracts/openapi/openapi.yaml (pure, auth-free)
##   Output: api/contracts/openapi/dist/openapi.bundled.yaml
openapi-bundle: openapi-lint
	@echo "[OPENAPI] Bundling api/contracts/openapi/openapi.yaml..."
	@mkdir -p api/contracts/openapi/dist
	@redocly bundle api/contracts/openapi/openapi.yaml -o api/contracts/openapi/dist/openapi.bundled.yaml
	@echo "[OPENAPI] Bundle complete -> api/contracts/openapi/dist/openapi.bundled.yaml"

## openapi-gateway: Validate and apply the gateway overlay to the bundled spec (Swagger UI only, never codegen)
##   Input:   api/contracts/openapi/dist/openapi.bundled.yaml (pure, auth-free)
##   Overlay: api/contracts/openapi/overlays/gateway.overlay.yaml
##   Output:  api/contracts/openapi/dist/openapi.gateway.yaml
openapi-gateway: openapi-bundle
	@echo "[OPENAPI] Applying gateway overlay..."
	@set -e; \
	if [ ! -x "$(OPENAPI_OVERLAY_BIN)" ]; then \
		echo "[OPENAPI][ERROR] overlay CLI not found at $(OPENAPI_OVERLAY_BIN). Run 'make install-tools' first." >&2; \
		exit 1; \
	fi; \
	"$(OPENAPI_OVERLAY_BIN)" overlay validate --overlay api/contracts/openapi/overlays/gateway.overlay.yaml; \
	"$(OPENAPI_OVERLAY_BIN)" overlay apply --overlay api/contracts/openapi/overlays/gateway.overlay.yaml --schema api/contracts/openapi/dist/openapi.bundled.yaml --out api/contracts/openapi/dist/openapi.gateway.yaml
	@echo "[OPENAPI] Gateway overlay complete -> api/contracts/openapi/dist/openapi.gateway.yaml"

## openapi-gen: Bundle the unified platform contract then generate the single Go SDK
##   Output: platform/platform-sdk/go/gen/openapi/platform/v1/platform.gen.go (package platformv1)
##   NOTE: consumes pure dist/openapi.bundled.yaml only -- never dist/openapi.gateway.yaml
openapi-gen: openapi-bundle
	@echo "[OPENAPI] Generating Go SDK from api/contracts/openapi/dist/openapi.bundled.yaml..."
	@mkdir -p platform/platform-sdk/go/gen/openapi/platform/v1
	@oapi-codegen -package platformv1 -generate types,client,chi-server,spec \
		-o platform/platform-sdk/go/gen/openapi/platform/v1/platform.gen.go api/contracts/openapi/dist/openapi.bundled.yaml
	@echo "[OPENAPI] Success -> platform/platform-sdk/go/gen/openapi/platform/v1/platform.gen.go"

# ==============================================================================
# Docker Compose (split files)
# ==============================================================================
# docker-compose.yml            = infra only (traefik, db, kafka, ...)
# docker-compose.services.yml   = domain services overlay (opt-in)
# docker-compose.docs.yml       = swagger-ui overlay (opt-in, Traefik-ready)

## infra-up: Start backing infrastructure only (never app services or docs)
infra-up:
	@docker compose $(COMPOSE_BASE) up -d

## infra-down: Stop backing infrastructure
infra-down:
	@docker compose $(COMPOSE_BASE) down

## infra-ps: Check running infrastructure container status
infra-ps:
	@docker compose $(COMPOSE_BASE) ps

## infra-logs: Follow logs from infrastructure containers
infra-logs:
	@docker compose $(COMPOSE_BASE) logs -f

## services-up: Build and start containerized domain services (on top of infra)
services-up:
	@docker compose $(COMPOSE_SERVICES) up -d --build

## services-down: Stop containerized domain services (infra keeps running)
services-down:
	@docker compose $(COMPOSE_SERVICES) stop production-service || true

## services-logs: Follow logs from containerized domain services
services-logs:
	@docker compose $(COMPOSE_SERVICES) logs -f production-service

## docs-up: Bundle specs + apply gateway overlays then start dev-only Swagger UI (http://localhost:3080/docs)
docs-up: openapi-gateway
	@docker compose $(COMPOSE_DOCS) up -d swagger-ui
	@echo "[DOCS] Swagger UI -> http://localhost:3080/docs"

## docs-down: Stop dev-only Swagger UI
docs-down:
	@docker compose $(COMPOSE_DOCS) stop swagger-ui || true

## docs-logs: Follow logs from dev-only Swagger UI
docs-logs:
	@docker compose $(COMPOSE_DOCS) logs -f swagger-ui

# ==============================================================================
# Cleanup
# ==============================================================================

## clean: Remove build artifacts and temporary binaries
clean:
	@echo "[CLEAN] Removing $(BIN_DIR) and test artifacts..."
	@rm -rf $(BIN_DIR) *.test *.out coverage.html
	@find . -name "coverage.out" -path "*/services/*" -o -name "coverage.out" -path "*/platform/*" | xargs rm -f
	@echo "[CLEAN] Done."
