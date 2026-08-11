# ==============================================================================
# FactoryOS Monorepo Makefile (Optional Developer Convenience Shortcuts)
# ==============================================================================

SHELL := /bin/bash
BIN_DIR := bin
GO_ENV := DATABASE_URL="postgres://factoryos:factoryos_password@localhost:5432/factoryos?sslmode=disable"

.PHONY: all help build build-all build-analytics build-edge build-simulator \
        test test-all test-analytics test-edge test-sdk test-coverage \
        run-analytics run-edge run-simulator \
        proto-lint proto-gen \
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
build-all: build-analytics build-edge build-simulator

## build-analytics: Build binary for Analytics Engine into $(BIN_DIR)/analytics-engine
build-analytics:
	@mkdir -p $(BIN_DIR)
	@echo "[BUILD] Compiling services/analytics-engine..."
	@go build -o $(BIN_DIR)/analytics-engine ./services/analytics-engine
	@echo "[BUILD] Success -> $(BIN_DIR)/analytics-engine"

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

# ==============================================================================
# Test Targets
# ==============================================================================

## test: Run unit tests across all Go modules
test: test-all

## test-all: Run all unit tests for Analytics Engine, Edge Runtime, and Platform SDK
test-all: test-edge test-sdk test-analytics

## test-analytics: Run unit tests for Analytics Engine (with coverage & race detector)
test-analytics:
	@echo "[TEST] Running tests for analytics-engine..."
	@cd services/analytics-engine && $(GO_ENV) go test -race -cover -v -timeout 60s ./processor/... ./db/... ./consumer/... ./config/...

## test-edge: Run unit tests for Edge Runtime
test-edge:
	@echo "[TEST] Running tests for edge-runtime..."
	@cd platform/edge-runtime && go test -v -cover ./...

## test-sdk: Run unit tests for Platform SDK
test-sdk:
	@echo "[TEST] Running tests for platform-sdk..."
	@cd platform/platform-sdk && go test -v ./...

## test-coverage: Run tests and generate statement coverage report for analytics-engine
test-coverage:
	@echo "[COVERAGE] Generating detailed code coverage report for analytics-engine..."
	@cd services/analytics-engine && $(GO_ENV) go test -race -coverprofile=coverage.out ./processor/... ./db/... ./consumer/... ./config/... && go tool cover -func=coverage.out

# ==============================================================================
# Run Targets
# ==============================================================================

## run-analytics: Build and execute Analytics Engine locally
run-analytics: build-analytics
	@echo "[RUN] Starting $(BIN_DIR)/analytics-engine..."
	@$(GO_ENV) $(BIN_DIR)/analytics-engine

## run-edge: Build and execute Edge Runtime locally
run-edge: build-edge
	@echo "[RUN] Starting $(BIN_DIR)/edge-runtime..."
	@$(BIN_DIR)/edge-runtime

## run-simulator: Execute Mock PLC Simulator locally
run-simulator:
	@echo "[RUN] Starting Mock PLC Simulator..."
	@cd examples/mock-plc-simulator && go run main.go

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
	@rm -rf $(BIN_DIR) *.test *.out services/analytics-engine/*.out platform/edge-runtime/*.out coverage.html
	@echo "[CLEAN] Done."
