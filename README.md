# FactoryOS

**The Event-Driven Manufacturing Operations Platform**

FactoryOS is an enterprise-grade, event-driven operational system of action designed to orchestrate modern manufacturing facilities. It bridges the gap between static enterprise resource planning (ERP) systems and real-time physical shop floor operations.

> **Mental Model:** FactoryOS combines the financial/domain context of SAP, the real-time physical ontology of Palantir, the high-throughput operational metrics of Datadog, and the developer-first extensibility of the Stripe platform.

---

## 🏛️ Platform Architecture & Blueprint

For full product strategy, technical architecture, and long-term evolutionary roadmap documents, inspect the `docs/` hierarchy:

* [Product Vision](docs/01-vision/PRODUCT_VISION.md)
* [System Overview](docs/02-architecture/SYSTEM_OVERVIEW.md)
* [Event-Driven Paradigm](docs/02-architecture/EVENT_DRIVEN_PARADIGM.md)
* [Architecture Decision Records (ADRs)](docs/03-adr/)
* [Async Event Catalog](docs/05-api/ASYNC_EVENT_CATALOG.md)
* [Strategic Roadmap](docs/06-roadmap/STRATEGIC_ROADMAP.md)
* [Developer Contributing Guide](docs/08-developer-guide/CONTRIBUTING.md)

---

## 📂 Monorepo Structure

```
factoryos/
├── docs/                 # Central product & engineering documentation hierarchy
├── api/                  # API contracts, schemas & architecture models
│   ├── contracts/        # Schema definitions (gRPC Protobuf, AsyncAPI, OpenAPI)
│   └── architecture/     # C4 diagrams, threat models, system specifications
├── platform/             # Core platform engines & shared frameworks
│   ├── workflow-engine/  # Durable state machine execution engine
│   ├── edge-runtime/     # Local factory edge agent (OPC-UA/MQTT ingestion & offline buffer)
│   └── platform-sdk/     # Go, TypeScript, Python client SDKs and context utilities
├── services/             # Core business domain microservices
│   ├── production-service/   # Work orders, routings, station execution, operator UI
│   ├── warehouse-service/    # Material staging, lot tracking, floor stock allocation
│   ├── maintenance-service/  # CMMS, asset health monitoring, PM schedules, spare parts
│   ├── quality-service/      # Inline inspection, SPC, Non-Conformance Reports (NCR)
│   ├── planning-service/     # Finite capacity dispatching & scheduling optimization
│   └── analytics-engine/     # Time-series telemetry aggregation (OEE, TEEP, MTBF, MTTR)
├── deploy/               # Infrastructure & deployment manifests
│   ├── terraform/        # Infrastructure-as-Code (Terraform/OpenTofu, Terragrunt)
│   └── helm/             # Kubernetes cloud control plane & edge cluster charts
└── examples/             # Reference integrations, mock PLC simulators, SDK usage
```

---

## 🗺️ Evolution Milestones

- **Milestone 1:** Telemetry Ingestion & Real-Time OEE Observability Core
- **Milestone 2:** Guided Digital Execution, Operator Terminal & Lot Tracking
- **Milestone 3:** Reactive Distributed Workflows (Quality NCR & Maintenance Auto-Dispatch)
- **Milestone 4:** Finite Capacity Dispatching & Dynamic Station Re-Routing
- **Milestone 5:** Enterprise Multi-Plant Control Tower & Open Developer Ecosystem

---

## 📜 License & Governance

Internal Enterprise Blueprint — FactoryOS Proprietary Platform Architecture.
