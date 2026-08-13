# ==============================================================================
# FactoryOS Monorepo Makefile (Optional Developer Convenience Shortcuts)
# ==============================================================================

SHELL := /bin/bash
BIN_DIR := bin

.PHONY: all help build build-all build-analytics build-ingestion build-edge build-simulator build-resource \
        test test-all test-analytics test-ingestion test-edge test-sdk test-resource \
        test-coverage test-coverage-analytics test-coverage-resource test-coverage-ingestion test-coverage-edge test-coverage-sdk \
        run-analytics run-ingestion run-edge run-simulator run-resource \
        proto-lint proto-gen openapi-bundle openapi-gen \
        infra-up infra-down infra-ps infra-logs clean

all: help

## help: Display available commands
help:
	@echo "FactoryOS Monorepo - Available commands (Optional Convenience):"
	@sed -n "s/^##//p" $(MAKEFILE_LIST) | column -t -s ":" | sed -e "s/^/ /"

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

## openapi-bundle: Bundle all multi-file OpenAPI domain contracts into dist/ (via Redocly)
##   Auto-discovers every api/contracts/openapi/<domain>/<version>/openapi.yaml
openapi-bundle:
	@echo "[OPENAPI] Bundling all domain specs in api/contracts/openapi/..."
	@find api/contracts/openapi -name "openapi.yaml" -not -path "*/dist/*" | while read spec; do \
		dir=$$(dirname $$spec); \
		mkdir -p $$dir/dist; \
		echo "  [BUNDLE] $$spec"; \
		redocly bundle $$spec -o $$dir/dist/openapi.bundled.yaml 2>/dev/null; \
	done
	@echo "[OPENAPI] Bundle complete -> api/contracts/openapi/**/dist/openapi.bundled.yaml"

## openapi-gen: Bundle all OpenAPI domain contracts then generate Go SDKs for each
##   Output: platform/platform-sdk/go/gen/openapi/<domain>/<version>/<domain>.gen.go
openapi-gen: openapi-bundle
	@echo "[OPENAPI] Generating Go SDKs from all bundled domain specs..."
	@find api/contracts/openapi -name "openapi.bundled.yaml" | while read bundle; do \
		version_dir=$$(dirname $$bundle | sed 's|/dist$$||'); \
		domain=$$(echo $$version_dir | awk -F'/' '{print $$(NF-1)}'); \
		version=$$(echo $$version_dir | awk -F'/' '{print $$NF}'); \
		pkg=$${domain}$${version}; \
		out_dir=platform/platform-sdk/go/gen/openapi/$$domain/$$version; \
		out_file=$$out_dir/$$domain.gen.go; \
		mkdir -p $$out_dir; \
		echo "  [GEN] $$domain/$$version -> $$out_file (package: $$pkg)"; \
		oapi-codegen -package $$pkg -generate types,client,chi-server,spec \
			-o $$out_file $$bundle; \
	done
	@echo "[OPENAPI] Success -> platform/platform-sdk/go/gen/openapi/"

# ==============================================================================
# Infrastructure Helpers (Docker Compose)
# ==============================================================================

## infra-up: Start all backing infrastructure containers in background
infra-up:
	@docker compose up -d

## infra-down: Stop all backing infrastructure containers
infra-down:
	@docker compose down

## infra-ps: Check running infrastructure container status
infra-ps:
	@docker compose ps

## infra-logs: Follow logs from infrastructure containers
infra-logs:
	@docker compose logs -f

# ==============================================================================
# Cleanup
# ==============================================================================

## clean: Remove build artifacts and temporary binaries
clean:
	@echo "[CLEAN] Removing $(BIN_DIR) and test artifacts..."
	@rm -rf $(BIN_DIR) *.test *.out coverage.html
	@find . -name "coverage.out" -path "*/services/*" -o -name "coverage.out" -path "*/platform/*" | xargs rm -f
	@echo "[CLEAN] Done."
