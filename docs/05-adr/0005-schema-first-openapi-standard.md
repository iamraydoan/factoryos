# ADR-0005: Schema-First OpenAPI 3.0/3.1 Standard for RESTful Interfaces

* **Status:** Accepted
* **Date:** 2026-08-11
* **Authors:** FactoryOS Architecture Team
* **Deciders:** Principal Architect & Lead Engineers

---

## 1. Context & Problem Statement

FactoryOS utilizes three distinct communication paradigms across its microservice ecosystem:
1. **IoT Ingestion & IPC:** High-throughput streaming via **gRPC (HTTP/2 + Protobuf)**.
2. **Asynchronous Messaging:** Cross-service domain event streaming via **Apache Kafka (AsyncAPI)**.
3. **User Interfaces & Enterprise Integration:** Synchronous RESTful APIs for Web UI Dashboards (Next.js/React), Mobile Supervisor Apps, and external ERP/MES systems (SAP, Oracle, SCADA).

Without a strictly enforced contract standard for RESTful endpoints, backend and frontend teams risk schema divergence (drift), lack of type safety, outdated documentation, and manual client SDK authoring.

---

## 2. Decision Drivers

* **Schema-First Mandate:** Strict alignment with [PROJECT_BIBLE.md](file:///workspaces/factoryos/docs/00-governance/PROJECT_BIBLE.md) rule #2 (Never code without schema contracts).
* **End-to-End Type Safety:** Automatic generation of Go server/client interfaces and TypeScript frontend types.
* **Interactive Living Documentation:** Self-hosted, embedded Swagger UI / Scalar endpoints without external internet dependencies.
* **Tooling Maturity:** Seamless integration with Go toolchains via `oapi-codegen/v2`.

---

## 3. Considered Options

* **Option A: Code-First Annotations (`swaggo/swag`)**
  * *Pros:* Fast to annotate on Go handlers.
  * *Cons:* Violates Schema-First mandate; schemas drift easily from actual API logic; cannot generate frontend types before backend implementation.
* **Option B: Schema-First with OpenAPI 3.0/3.1 & `oapi-codegen` (Chosen)**
  * *Pros:* 100% Schema-First; Single Source of Truth under `api/contracts/openapi/`; generates Go DTOs and Chi/HTTP server interfaces into `platform-sdk`; enables `openapi-typescript` for frontend.
  * *Cons:* Requires defining YAML specs before coding handlers (desired governance pattern).

---

## 4. Decision

We adopt **Option B: Schema-First OpenAPI 3.0/3.1 with `oapi-codegen`**.

### Specifications & Directory Layout:
```
api/contracts/
├── telemetry/v1/ingestion.proto              # gRPC Protobuf Contract
├── resource/v1/equipment.proto               # gRPC Protobuf Contract
└── openapi/                                  # OpenAPI 3.0/3.1 Specifications
    ├── telemetry/v1/openapi.yaml             # Telemetry & Analytics Query REST API
    ├── resource/v1/openapi.yaml              # Resource Management & Master Data REST API
    └── production/v1/openapi.yaml            # Work Order & Execution REST API
```

### Automation & Tooling:
* `make openapi-gen`: Compiles OpenAPI contracts into `platform/platform-sdk/go/gen/openapi/...`.
* [swaggerui](file:///workspaces/factoryos/platform/platform-sdk/go/swaggerui/swaggerui.go): Standard embedded HTTP handler serving Scalar/Swagger documentation on `/docs` or `/swagger`.

---

## 5. Consequences

### Positive:
* **Zero Schema Drift:** Backend handlers implement generated Go interfaces; frontend builds from the same OpenAPI YAML.
* **Self-Documenting Services:** Every microservice serves interactive documentation out-of-the-box.
* **Enterprise Integration Ready:** Third-party factory IT teams can import `/openapi.yaml` directly into Postman or API Gateways.

### Negative / Mitigations:
* Requires developers to learn OpenAPI 3.0/3.1 syntax. Mitigated by providing clear templates and `make openapi-gen` validation in CI.
