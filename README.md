# FactoryOS

**The Event-Driven Manufacturing Operations Platform**

FactoryOS is an enterprise-grade, event-driven operational system of action designed to orchestrate modern manufacturing facilities. It bridges the gap between static enterprise resource planning (ERP) systems and real-time physical shop floor operations.

> **Mental Model:** FactoryOS combines the financial/domain context of SAP, the real-time physical ontology of Palantir, the high-throughput operational metrics of Datadog, and the developer-first extensibility of the Stripe platform.

---

## 🏛️ Platform Architecture & Governance

For full product strategy, technical architecture, governance standards, and long-term evolutionary roadmap documents, inspect the `docs/` hierarchy:

* [Project Bible & Governance](docs/00-governance/PROJECT_BIBLE.md)
* [Product Overview & Vision](docs/01-overview/PRODUCT_VISION.md)
* [Business Processes & User Roles](docs/02-business/)
* [Domain Model & Bounded Contexts](docs/03-domain/)
* [System Architecture & C4 Diagrams](docs/04-architecture/)
* [Architecture Decision Records (ADRs)](docs/05-adr/)
* [Request for Comments (RFCs)](docs/06-rfc/)
* [Async Event Catalog](docs/07-api/ASYNC_EVENT_CATALOG.md)
* [Strategic Roadmap & Milestones](docs/08-roadmap/MILESTONES.md)
* [Task Epics](docs/09-epics/)
* [Developer Contributing Guide](docs/10-developer-guide/CONTRIBUTING.md)

---

## 📂 Monorepo Structure

```
factoryos/
├── docs/                 # Central product & engineering documentation hierarchy
│   ├── 00-governance/    # Project Bible & core engineering standards
│   ├── 01-overview/      # Product vision, value proposition, glossary
│   ├── 02-business/      # Business processes, user roles & user stories
│   ├── 03-domain/        # Bounded contexts (ISA-95, Production, Genealogy, Quality, CMMS)
│   ├── 04-architecture/  # C4 diagrams & system overview
│   ├── 05-adr/           # Architecture Decision Records
│   ├── 06-rfc/           # Feature & capability RFC proposals
│   ├── 07-api/           # Async event catalog & API specs
│   ├── 08-roadmap/       # Evolution milestones & strategic roadmap
│   ├── 09-epics/         # Development task epics & execution checklists
│   └── 10-developer-guide/# Developer contributing guide & onboarding

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

> See full detail in [docs/07-roadmap/MILESTONES.md](docs/07-roadmap/MILESTONES.md) and [docs/07-roadmap/STRATEGIC_ROADMAP.md](docs/07-roadmap/STRATEGIC_ROADMAP.md).

| Milestone | Focus | Horizon |
|---|---|---|
| **M1** | Physical Ontology & Real-Time OEE Telemetry | Months 0–6 |
| **M2** | Guided Digital Execution & Lot Genealogy | Months 6–12 |
| **M3** | Closed-Loop Quality NCR & CMMS Maintenance | Months 12–18 |
| **M4** | Finite Capacity Scheduling & Edge Resilience | Months 18–24 |
| **M5** | Multi-Plant Control Tower & Industrial AI | Months 24–36 |


---

## 📜 License & Governance

Internal Enterprise Blueprint — FactoryOS Proprietary Platform Architecture.
