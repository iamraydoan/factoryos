# Local Environment Setup Guide

This document provides instructions for new developers to spin up the local FactoryOS infrastructure. As we add new dependencies or services, please update this guide.

---

## 1. Prerequisites

Before starting, ensure you have the following installed on your machine:
* **Docker Desktop** (or Docker Engine + Docker Compose)
* **Git**
* At least **8GB RAM** allocated to Docker (the infra stack runs multiple databases and JVMs).

---

## 1.1 Development Environment Options

Developers have **two choices** for setting up their workspace. Dev Container is **optional** but recommended.

### Option A: Dev Container (Recommended - Zero Local Tool Setup)
If you use VS Code or cursor:
1. Ensure Docker Desktop is running.
2. Open the project in VS Code.
3. Click **"Reopen in Container"** when prompted (or `Cmd+Shift+P` -> `Dev Containers: Reopen in Container`).
4. VS Code will spin up a pre-configured Ubuntu container with **Java 21, Go, Node.js, and Protobuf plugins** pre-installed.
5. The container automatically runs `make install-tools` on create so the Buf, Protobuf, OpenAPI, and Redocly toolchain is ready to use.

### Option B: Native Host Machine (Manual Setup)
If you prefer coding directly on your Mac/Linux/Windows machine without Dev Containers, install the required toolchain first and then run the repo's single setup command:
* **JDK 21** (Amazon Corretto, Temurin, or Zulu)
* **Go 1.22+**
* **Node.js 20+**
* **Make** (Optional - for shortcut targets)

Then run:
```bash
make install-tools
```

This is the same toolchain the devcontainer installs automatically during container creation. The command is the canonical source of truth; do not treat the block below as an additional second installation step.

```bash
# Reference only: same commands executed by `make install-tools`
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/bufbuild/buf/cmd/buf@latest
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1
npm install --global --no-fund --no-audit @redocly/cli@1.34.0
```

---

## 2. Spinning Up the Infrastructure

FactoryOS relies on a comprehensive local infrastructure stack (Kafka, Postgres, Zitadel, Valkey). To start the entire stack:

1. Open your terminal at the root of the `factoryos` project.
2. (Optional) Configure local infrastructure ports and credentials. Defaults work without this file:

```bash
cp .env.example .env
```

Edit `.env` before starting the stack if any default host port is already in use. The sample credentials and Zitadel master key are for local development only.

3. Start the infrastructure:

```bash
docker compose up -d
```

4. (First time only) Docker will pull all the latest images. This may take a few minutes depending on your internet connection.
5. Verify all containers are running and healthy:

```bash
docker compose ps
```

---

## 3. Local Service Directory

Once the stack is up, the following services and ports are available on your `localhost`:

| Service | Port | Description | Credentials / Access |
|---|---|---|---|
| **Traefik Dashboard** | `8080` (`TRAEFIK_DASHBOARD_PORT`) | API Gateway routing UI | http://localhost:8080 |
| **Traefik Ingress** | `80` (`TRAEFIK_HTTP_PORT`) | Main entrypoint for HTTP requests | http://localhost |
| **Traefik gRPC** | `50051` (`TRAEFIK_GRPC_PORT`) | Main entrypoint for gRPC requests | `localhost:50051` |
| **FactoryOS DB** | `5432` (`FACTORYOS_DB_PORT`) | TimescaleDB for core services & telemetry | User: `factoryos` / Pass: `factoryos_password` / DB: `factoryos` |
| **Zitadel DB** | `5433` (`ZITADEL_DB_PORT`) | Dedicated Postgres for IAM | User: `postgres` / Pass: `zitadel_password` / DB: `zitadel` |
| **Kafka (KRaft)** | `9092` (`KAFKA_PORT`) | Event Bus broker | `localhost:9092` |
| **Valkey (Cache)** | `6379` (`VALKEY_PORT`) | Redis drop-in replacement | `localhost:6379` |
| **Mosquitto MQTT** | `1883` (`MQTT_PORT`) | Edge MQTT Broker for IIoT telemetry | `localhost:1883` |
| **Zitadel Console** | `8081` (`ZITADEL_PORT`) | IAM Web Interface | http://localhost:8081/ui/console (User: `zitadel-admin@zitadel.localhost` / Pass: `Password123!`) |
| **Swagger UI (docs profile, opt-in)** | `8082` (`SWAGGER_UI_PORT`) | Dev/tester-only OpenAPI viewer (telemetry, no Traefik) | http://localhost:8082/docs — start with `make docs-up` after `make openapi-bundle` |

---

## 4. Troubleshooting & Useful Commands

**View logs for all services:**
```bash
docker compose logs -f
```

**View logs for a specific service (e.g., Kafka):**
```bash
docker compose logs -f kafka
```

**Shut down the infrastructure:**
```bash
docker compose down
```

**Completely wipe database data (Use with caution!):**
```bash
docker compose down -v
```

---

## 5. Java Services Configuration

Java services (Production, Warehouse, Quality, Maintenance) use environment variables for configuration.

### Environment Files

Connection details and secrets live in env files only — never in `application-*.yml`
profiles. Profiles carry behavior flags (e.g. `show-sql`) while hosts and
credentials come from the environment.

| File | Purpose |
|------|---------|
| Root `.env.example` / `.env.docker.example` | Docker Compose ports and local infrastructure credentials (host-run vs container network) |
| `services/<service>/.env.example` | Service-specific template for host-local runs (`DB_HOST=localhost`) |
| `services/<service>/.env.docker.example` | Service-specific template for runs inside the Docker network (`DB_HOST=factoryos-db`) |

### Setup

```bash
# Optional: Docker Compose infrastructure settings, from the repository root
cp .env.example .env

# Java service settings remain service-specific — pick the template
# matching where the service process runs:
cd services/production-service
cp .env.example .env                 # host machine (DB at localhost)
cp .env.docker.example .env         # devcontainer / Docker network (DB at factoryos-db)
```

### Spring Profiles

The production-service uses Spring profiles for behavior flags only:

| Profile | Use Case |
|---------|----------|
| *(none)* | Default — works out of the box, connection from env (`DB_HOST`, default `localhost`) |
| `local` | Local dev with `show-sql` enabled |

### Running Java Services

Spring reads OS env, not `.env` files — always run via `make` so the
`.env` is sourced first. Raw `mvn spring-boot:run` ignores `.env` and
falls back to `localhost`.

```bash
# Host machine (DB at localhost)
make setup-production-host
make run-production

# With SQL logging
make run-production-local

# Inside devcontainer / Docker network (DB at factoryos-db)
make setup-production-docker
make run-production
```

---

## 6. Testing Telemetry & Mock PLC Simulation

To test end-to-end telemetry ingestion locally:

1. **Start MQTT Broker:**
   ```bash
   docker compose up -d mosquitto
   ```

2. **Start Edge Runtime (Ingestion & SQLite Buffer):**
   ```bash
   cd platform/edge-runtime
   go run main.go
   ```

3. **Start Mock PLC Simulator (In a second terminal):**
   ```bash
   cd examples/mock-plc-simulator
   go run main.go
   ```

*(See full simulator configuration guide in [examples/README.md](../../examples/README.md)).*

---

## 7. Makefile Shortcuts (Optional)

For developers who prefer using `make`, a top-level `Makefile` is provided with convenient shortcuts. Running commands directly via `go` or `docker compose` is always supported.

| Task | Make Shortcut | Direct Command Equivalent |
|---|---|---|
| **View help** | `make help` | — |
| **Install dev toolchain** | `make install-tools` | `go install ...` + `npm install --global @redocly/cli` |
| **Build all binaries** | `make build` | `go build ./...` |
| **Build specific service** | `make build-analytics`<br>`make build-edge`<br>`make build-simulator` | `go build -o bin/<service> ./...` |
| **Run all Go tests** | `make test` | `go test ./...` |
| **Test Analytics Engine** | `make test-analytics` | `cd services/analytics-engine && go test -race -cover -v ./...` |
| **Coverage Report** | `make test-coverage` | `go test -coverprofile=... && go tool cover -func=...` |
| **Run Service Locally** | `make run-analytics`<br>`make run-edge`<br>`make run-simulator`<br>`make run-production`<br>`make run-production-local` | `go run main.go`<br>`mvn spring-boot:run` (with `.env` sourced) |
| **Protobuf Lint / Gen** | `make proto-lint`<br>`make proto-gen` | `cd api/contracts && buf lint`<br>`cd api/contracts && buf generate` |
| **Docker Infra** | `make infra-up`<br>`make infra-down`<br>`make infra-logs` | `docker compose up -d`<br>`docker compose down`<br>`docker compose logs -f` |
| **Cleanup** | `make clean` | `rm -rf bin/ *.out` |

---

## 8. CI/CD & Automated Release Pipelines

The repository uses GitHub Actions workflows for continuous integration and automated binary releases:

1. **Continuous Integration (`.github/workflows/ci.yml`):**
   - Automatically runs unit tests on all Pull Requests and pushes to `main`.
   - Enforces a strict **Code Coverage Gate (Coverage >= 80%)**. Fails PR builds if total coverage drops below threshold.

2. **Automated Multi-Arch Releases (`.github/workflows/release.yml`):**
   - Automatically triggers on new version tags (e.g., `git tag v0.1.0 && git push --tags`).
   - Cross-compiles `edge-runtime` for 4 platforms: `linux-amd64`, `linux-arm64`, `windows-amd64.exe`, and `darwin-arm64`.
   - Attaches packaged binaries directly to the GitHub Release page.

---

## 9. Adding New Services (For Maintainers)

When adding a new backing service (e.g., Temporal, OpenTelemetry) to `docker-compose.yml`:
1. Ensure you use a **specific image version tag** (avoid `latest`).
2. Add a **named volume** if the service requires persistent state.
3. Update the "Local Service Directory" table above so the team knows the new ports.

> **Docs profile:** `swagger-ui` runs under Compose profile `docs` and is not part of
> default `docker compose up -d`. Use `make docs-up` / `make docs-down`.
> Prerequisite: `make openapi-bundle` (the `dist/*.bundled.yaml` output is gitignored).
