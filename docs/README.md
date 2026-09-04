# FactoryOS Documentation Directory & Index

> Welcome to the official documentation repository for **FactoryOS** — an enterprise event-driven manufacturing operations platform.

---

## 🌲 Documentation Tree Index

```
docs/
├── 📁 00-governance/              # System Governance & Standards
│   └── 📄 PROJECT_BIBLE.md        # The North Star for Engineering Standards & Rules
│
├── 📁 01-overview/                # Product Vision & High-Level Overview
│   ├── 📄 PRODUCT_VISION.md       # Product Vision & Strategic Positioning
│   ├── 📄 VALUE_PROPOSITION.md    # Business Value Proposition
│   └── 📄 TERMINOLOGY_GLOSSARY.md # Core System & Industry Glossary
│
├── 📁 02-business/                # Business Context & Specifications
│   ├── 📄 Business-Glossary.md    # Manufacturing Business Terminology
│   ├── 📄 Business-Processes.md   # Core Manufacturing Business Workflows
│   ├── 📄 User-Roles.md           # User Personas, Roles & Access Control
│   └── 📄 User-Stories.md         # User Epics & Core Feature Stories
│
├── 📁 03-domain/                  # ISA-95 & Domain Bounded Contexts
│   ├── 📄 00-Domain-Overview.md   # Overview of Bounded Contexts & Domain Architecture
│   ├── 📁 01-resource-management/
│   │   ├── 📄 Equipment-Hierarchy.md
│   │   ├── 📄 Material-Definition.md
│   │   └── 📄 Personnel.md
│   ├── 📁 02-production-operations/
│   │   ├── 📄 Production-Request.md
│   │   ├── 📄 Production-Schedule.md
│   │   ├── 📄 Work-Order.md
│   │   ├── 📄 Operation.md
│   │   ├── 📄 Execution.md
│   │   └── 📄 Performance.md
│   ├── 📁 03-quality-operations/
│   │   ├── 📄 Test-Specification.md
│   │   ├── 📄 Inspection.md
│   │   ├── 📄 Non-Conformance.md
│   │   ├── 📄 Disposition.md
│   │   └── 📄 CAPA.md
│   ├── 📁 04-maintenance-operations/
│   │   ├── 📄 Physical-Asset.md
│   │   ├── 📄 Maintenance-Request.md
│   │   ├── 📄 PM-Schedule.md
│   │   ├── 📄 Work-Order.md
│   │   └── 📄 Downtime.md
│   └── 📁 05-inventory-operations/
│       ├── 📄 Material-Lot.md
│       ├── 📄 Material-Movement.md
│       └── 📄 Genealogy.md
│
├── 📁 04-architecture/            # Technical Architecture & C4 Models
│   ├── 📄 SYSTEM_OVERVIEW.md      # High-Level Architecture Overview
│   ├── 📄 C4-Context.md           # C4 Level 1 - System Context Diagram
│   ├── 📄 C4-Container.md         # C4 Level 2 - Container Diagram
│   └── 📄 EVENT_DRIVEN_PARADIGM.md# Event Classification & Outbox Pattern Specs
│
├── 📁 05-adr/                     # Architecture Decision Records
│   ├── 📄 0000-adr-template.md    # ADR Standard Template
│   ├── 📄 0001-use-event-driven-architecture.md
│   ├── 📄 0002-schema-first-api-contracts.md
│   └── 📄 0003-durable-execution-engine.md
│
├── 📁 06-rfc/                     # Request for Comments (Feature Specs)
│   ├── 📄 0000-rfc-template.md    # RFC Standard Template
│   ├── 📄 0001-asset-telemetry-ingestion.md
│   ├── 📄 0002-work-order-execution-engine.md
│   └── 📄 0003-material-genealogy-tracking.md
│
├── 📁 07-api/                     # API & Event Contracts Catalog
│   ├── 📄 ASYNC_EVENT_CATALOG.md  # Async Kafka Event Specification Catalog
│   └── 📄 PAGINATION_DESIGN.md   # Pagination Standard (gRPC cursor / REST page)
│
├── 📁 08-roadmap/                 # Evolutionary Roadmap & Milestones
│   └── 📄 MILESTONES.md           # Milestone 1 to 5 Deliverables & Timelines
│
├── 📁 09-epics/                   # Engineering Tasks & Checklists
│   ├── 📄 EPIC-001-Platform-Foundation.md
│   ├── 📄 EPIC-002-Resource-Management-Telemetry.md
│   ├── 📄 EPIC-003-Production-Operations.md
│   ├── 📄 EPIC-004-Inventory-Genealogy.md
│   ├── 📄 EPIC-005-Quality-Operations.md
│   └── 📄 EPIC-006-Maintenance-Operations.md
│
└── 📁 10-developer-guide/         # Developer Onboarding & Guides
    ├── 📄 Local-Environment-Setup.md # Docker Infrastructure & Dev Container Setup
    └── 📄 CONTRIBUTING.md         # Developer Contribution Guidelines
```

---

## 🔗 Quick Links

* [Project Bible & Governance](./00-governance/PROJECT_BIBLE.md)
* [Strategic Milestones & Roadmap](./08-roadmap/MILESTONES.md)
* [Active Epic: EPIC-001 Platform Foundation](./09-epics/EPIC-001-Platform-Foundation.md)
* [Local Environment Setup](./10-developer-guide/Local-Environment-Setup.md)
* [Pagination Design Standard](./07-api/PAGINATION_DESIGN.md)

---

## 🛠️ Development Lifecycle Workflow

1. **Check Strategy:** Refer to [`./08-roadmap/MILESTONES.md`](./08-roadmap/MILESTONES.md).
2. **Review RFC & ADR:** Ensure spec exists in [`./06-rfc/`](./06-rfc) and decision is logged in [`./05-adr/`](./05-adr).
3. **Review Epic Tasks:** Check checklist in [`./09-epics/`](./09-epics).
4. **Schema-First Specification:** Define Protobuf/AsyncAPI in `api/contracts/`.
5. **Implementation & Check:** Write code, pass unit tests, and mark task `[x]` in Epic.
