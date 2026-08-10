# ==============================================================================
# FactoryOS Monorepo Makefile
# ==============================================================================

SHELL := /bin/bash
BIN_DIR := bin
GO_ENV := DATABASE_URL="postgres://factoryos:factoryos_password@localhost:5432/factoryos?sslmode=disable"

.PHONY: all help build build-analytics test test-analytics test-coverage run-analytics clean

all: help

## help: Display available commands
help:
	@echo "FactoryOS Monorepo - Available commands:"
	@sed -n "s/^##//p" $(MAKEFILE_LIST) | column -t -s ":" | sed -e "s/^/ /"

# ==============================================================================
# Build Targets
# ==============================================================================

## build: Build all Go services into $(BIN_DIR)/
build: build-analytics

## build-analytics: Build binary for Analytics Engine into $(BIN_DIR)/analytics-engine
build-analytics:
	@mkdir -p $(BIN_DIR)
	@echo "[BUILD] Compiling services/analytics-engine..."
	@go build -o $(BIN_DIR)/analytics-engine ./services/analytics-engine
	@echo "[BUILD] Success -> $(BIN_DIR)/analytics-engine"

# ==============================================================================
# Test Targets
# ==============================================================================

## test: Run unit tests with race detector and coverage across services
test: test-analytics

## test-analytics: Run unit tests for Analytics Engine (with coverage & race detector)
test-analytics:
	@echo "[TEST] Running tests with race detector & coverage..."
	@cd services/analytics-engine && $(GO_ENV) go test -race -cover -v -timeout 60s ./processor/... ./db/... ./consumer/... ./config/...

## test-coverage: Run tests and generate statement coverage report
test-coverage:
	@echo "[COVERAGE] Generating detailed code coverage report..."
	@cd services/analytics-engine && $(GO_ENV) go test -race -coverprofile=coverage.out ./processor/... ./db/... ./consumer/... ./config/... && go tool cover -func=coverage.out

# ==============================================================================
# Run Targets
# ==============================================================================

## run-analytics: Build and execute Analytics Engine locally
run-analytics: build-analytics
	@echo "[RUN] Starting $(BIN_DIR)/analytics-engine..."
	@$(BIN_DIR)/analytics-engine

# ==============================================================================
# Cleanup
# ==============================================================================

## clean: Remove build artifacts and temporary binaries
clean:
	@echo "[CLEAN] Removing $(BIN_DIR) and test artifacts..."
	@rm -rf $(BIN_DIR) *.test *.out services/analytics-engine/*.out coverage.html
	@echo "[CLEAN] Done."
