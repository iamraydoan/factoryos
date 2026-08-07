# Changelog

All notable changes to the **FactoryOS** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Future enhancements and backlog ideas tracked in [INBOX.md](INBOX.md).

---

## [v0.1.0] - 2026-08-07

### Added
- **Edge Runtime Store-and-Forward SQLite Buffer (`platform/edge-runtime/buffer`):**
  - Offline-resilient SQLite local outbox table with WAL mode for high-concurrency event buffering.
  - Support for `Enqueue`, `DequeueBatch`, `MarkSent`, `IncrementRetryCount`, and `GetPendingCount`.
  - Serialized Protobuf payload storage (`telemetryv1.TelemetryPayload`).

- **Real-Time Telemetry Collector & Sync Worker (`platform/edge-runtime/collector`):**
  - Collector engine for packaging sensor metrics.
  - Background `SyncWorker` for periodic cloud flushing with backoff error retention during network outages.

- **Asynchronous MQTT Subscriber (`platform/edge-runtime/mqtt`):**
  - Non-blocking MQTT subscriber client (`paho.mqtt.golang`) listening on `factoryos/telemetry/+/readings`.
  - Dual payload decoder supporting both Protobuf binaries and JSON fallback formats.

- **Configurable Mock PLC Telemetry Simulator (`examples/mock-plc-simulator`):**
  - Machine simulator generating real-time telemetry (temperatures, vibration, production part counts, operational states).
  - External JSON configuration support via [config.json](examples/mock-plc-simulator/config.json) or `-config` flag.

- **Automated GitHub Actions CI/CD Pipelines:**
  - **CI Workflow (`.github/workflows/ci.yml`):** Automated testing, Go workspace compilation, and strict Code Coverage Gate (Threshold >= 80%).
  - **Release Workflow (`.github/workflows/release.yml`):** Cross-compilation of `edge-runtime` binaries across 4 platforms (`linux/amd64`, `linux/arm64`, `windows/amd64`, `darwin/arm64`) with automated GitHub Releases publishing.

- **Edge Development Infrastructure:**
  - Eclipse Mosquitto MQTT Broker integration in [docker-compose.yml](docker-compose.yml) listening on port 1883.
  - Developer Inbox & Parking Lot backlog file ([INBOX.md](INBOX.md)).
