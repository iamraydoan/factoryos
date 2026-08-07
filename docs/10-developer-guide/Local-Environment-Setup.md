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

### Option B: Native Host Machine (Manual Setup)
If you prefer coding directly on your Mac/Linux/Windows machine without Dev Containers, install these manually:
* **JDK 21** (Amazon Corretto, Temurin, or Zulu)
* **Go 1.22+**
* **Node.js 20+**
* **Protobuf Compiler (`protoc`)** & Go plugins:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  go install github.com/bufbuild/buf/cmd/buf@latest
  ```

---

## 2. Spinning Up the Infrastructure

FactoryOS relies on a comprehensive local infrastructure stack (Kafka, Postgres, Zitadel, Valkey). To start the entire stack:

1. Open your terminal at the root of the `factoryos` project.
2. Run the following command:

```bash
docker compose up -d
```

3. (First time only) Docker will pull all the latest images. This may take a few minutes depending on your internet connection.
4. Verify all containers are running and healthy:

```bash
docker compose ps
```

---

## 3. Local Service Directory

Once the stack is up, the following services and ports are available on your `localhost`:

| Service | Port | Description | Credentials / Access |
|---|---|---|---|
| **Traefik Dashboard** | `8080` | API Gateway routing UI | http://localhost:8080 |
| **Traefik Ingress** | `80` | Main entrypoint for HTTP requests | http://localhost |
| **Traefik gRPC** | `50051` | Main entrypoint for gRPC requests | `localhost:50051` |
| **FactoryOS DB** | `5432` | TimescaleDB for core services & telemetry | User: `factoryos` / Pass: `factoryos_password` / DB: `factoryos` |
| **Zitadel DB** | `5433` | Dedicated Postgres for IAM | User: `postgres` / Pass: `zitadel_password` / DB: `zitadel` |
| **Kafka (KRaft)** | `9092` | Event Bus broker | `localhost:9092` |
| **Valkey (Cache)** | `6379` | Redis drop-in replacement | `localhost:6379` |
| **Mosquitto MQTT** | `1883` | Edge MQTT Broker for IIoT telemetry | `localhost:1883` |
| **Zitadel Console** | `8081` | IAM Web Interface | http://localhost:8081/ui/console (User: `zitadel-admin@zitadel.localhost` / Pass: `Password123!`) |

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

## 5. Testing Telemetry & Mock PLC Simulation

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

## 6. Adding New Services (For Maintainers)

When adding a new backing service (e.g., Temporal, OpenTelemetry) to `docker-compose.yml`:
1. Ensure you use a **specific image version tag** (avoid `latest`).
2. Add a **named volume** if the service requires persistent state.
3. Update the "Local Service Directory" table above so the team knows the new ports.
