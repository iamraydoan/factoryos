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

FactoryOS relies on a comprehensive local infrastructure stack (Kafka, Postgres, Zitadel, Valkey). Compose is split into three files sharing `factoryos_net`:

| File | Contains | Started by |
|------|----------|------------|
| `docker-compose.yml` | Infra only (traefik, db, kafka, valkey, mosquitto) | `make infra-up` |
| `docker-compose.services.yml` | Containerized domain services (opt-in) | `make services-up` |
| `docker-compose.docs.yml` | Swagger UI (opt-in, Traefik-ready) | `make docs-up` |

`infra-up` never starts app services or docs. Run services on the host
via `make run-production` during development; use `services-up` only to
test the containerized build.

1. Open your terminal at the root of the `factoryos` project.
2. (Optional) Configure local infrastructure ports and credentials. Defaults work without this file:

```bash
cp .env.example .env
```

Edit `.env` before starting the stack if any default host port is already in use. The sample credentials and Zitadel master key are for local development only.

3. Start the infrastructure:

```bash
make infra-up
```

4. (First time only) Docker will pull all the latest images. This may take a few minutes depending on your internet connection.
5. Verify all containers are running and healthy:

```bash
make infra-ps
```

6. (Optional) Start containerized domain services or docs:

```bash
make services-up   # builds + starts production-service container
make docs-up       # bundles OpenAPI specs, starts Swagger UI on :3080/docs
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
| **Swagger UI (opt-in, `docker-compose.docs.yml`)** | `3080` (`SWAGGER_UI_PORT`) | Dev/tester-only OpenAPI viewer (telemetry, Traefik-ready) | http://localhost:3080/docs — start with `make docs-up` after `make openapi-bundle` |

### Domain Service Port Scheme

REST on **3xxx**, gRPC on **4xxx** (host = container). Infra keeps its
standard ports (`80/8080/50051/5432/...`). Every Java `application.yml`
uses `${SERVER_PORT:<default>}` / `${GRPC_SERVER_PORT:<default>}` — no
hardcoded ports.

| Service | REST | gRPC |
|---|---|---|
| production (Java) | `3001` | `4001` |
| quality (Java) | `3002` | `4002` |
| warehouse (Java) | `3003` | `4003` |
| maintenance (Java) | `3004` | `4004` |
| planning (Java) | `3005` | `4005` |
| resource (Go) | `3050` (health) | `4050` |
| ingestion (Go) | `3051` (metrics) | `4051` |
| analytics (Go) | `3052` (metrics) | — |

---

## 4. Troubleshooting & Useful Commands

**View logs for all infra services:**
```bash
make infra-logs
```

**View logs for a specific service (e.g., Kafka):**
```bash
docker compose -f docker-compose.yml logs -f kafka
```

**View containerized domain service logs:**
```bash
make services-logs
```

**Shut down the infrastructure:**
```bash
make infra-down
```

**Completely wipe database data (Use with caution!):**
```bash
docker compose -f docker-compose.yml down -v
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
| Root `.env.example` | **HOST** profile — Compose ports + credentials; gateway backends `host.docker.internal` |
| Root `.env.docker.example` | **DOCKER** profile — gateway backends as container names (`production-service`) |
| Root `.env.devcontainer.example` | **DEVCONTAINER** profile — gateway backends `host.docker.internal`; DB reachable by container name |
| `services/<service>/.env.example` | **HOST** — `DB_HOST=localhost` |
| `services/<service>/.env.docker.example` | **DOCKER** — `DB_HOST=factoryos-db` |
| `services/<service>/.env.devcontainer.example` | **DEVCONTAINER** — `DB_HOST=factoryos-db` (joins `factoryos_net`) |

### Setup

One command installs both files for a profile (root `.env` + service `.env`):

```bash
# Pick the profile matching where processes run:
make setup-env-host            # DB=localhost, gateway=host.docker.internal
make setup-env-docker          # DB + gateway = container names
make setup-env-devcontainer    # DB=factoryos-db, gateway=host.docker.internal
```

Manual equivalent (repository root):

```bash
cp .env.example .env                                    # root: Compose ports + credentials
cp services/production-service/.env.example services/production-service/.env
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
make setup-env-host
make run-production

# With SQL logging
make run-production-local

# Inside devcontainer / Docker network (DB at factoryos-db)
make setup-env-devcontainer
make run-production

# Fully containerized service (DB + gateway = container names)
make setup-env-docker
make services-up
```

---

## 6. Testing Telemetry & Mock PLC Simulation

To test end-to-end telemetry ingestion locally:

1. **Start MQTT Broker:**
   ```bash
   docker compose -f docker-compose.yml up -d mosquitto
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
| **Run all tests (Go + Java)** | `make test` / `make test-all` | `go test ./...` then `cd services && mvn verify` |
| **Run all Go tests only** | `make test-go` | `go test ./...` |
| **Run all Java tests only (+ coverage gate)** | `make test-java` | `cd services && mvn verify` |
| **Test one Java service** | `make test-java-service SERVICE=production-service` | `cd services && mvn -pl production-service verify` |
| **Java coverage report** | `make test-java` then open `services/*/target/site/jacoco/index.html` | `mvn verify` then open the JaCoCo HTML report |
| **Test Analytics Engine** | `make test-analytics` | `cd services/analytics-engine && go test -race -cover -v ./...` |
| **Coverage Report (Go)** | `make test-coverage` | `go test -coverprofile=... && go tool cover -func=...` |
| **Run Service Locally** | `make run-analytics`<br>`make run-edge`<br>`make run-simulator`<br>`make run-production`<br>`make run-production-local` | `go run main.go`<br>`mvn spring-boot:run` (with `.env` sourced) |
| **Protobuf Lint / Gen** | `make proto-lint`<br>`make proto-gen` | `cd api/contracts && buf lint`<br>`cd api/contracts && buf generate` |
| **Docker Infra** | `make infra-up`<br>`make infra-down`<br>`make infra-logs` | `docker compose -f docker-compose.yml up -d`<br>`docker compose -f docker-compose.yml down`<br>`docker compose -f docker-compose.yml logs -f` |
| **Docker Services** | `make services-up`<br>`make services-down`<br>`make services-logs` | `docker compose -f docker-compose.yml -f docker-compose.services.yml up -d --build`<br>`... stop production-service`<br>`... logs -f production-service` |
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

When adding a new domain service, put it in `docker-compose.services.yml`
(never in the infra file) so `infra-up` stays infra-only.

> **Docs overlay:** `swagger-ui` lives in `docker-compose.docs.yml` and is not part of
> `infra-up`. Use `make docs-up` / `make docs-down`. File inclusion is the opt-in
> (no Compose `profiles:`). The service stays on `factoryos_net` so a future
> Traefik route (`PathPrefix /docs` -> `http://swagger-ui:8080`) needs no infra change.
> Prerequisite: `make openapi-bundle` (the `dist/*.bundled.yaml` output is gitignored).
