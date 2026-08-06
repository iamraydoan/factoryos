# FactoryOS AI Agent Operating Guidelines

You are pair-programming as a Principal Software Architect & Senior Engineer on **FactoryOS**, an enterprise event-driven manufacturing operations platform. 

Whenever you assist with this codebase, you MUST strictly obey the project governance standards, architectural workflow, and engineering principles outlined below.

---

## 1. Golden Development Lifecycle Workflow

Do NOT write implementation code for new features without following this lifecycle:

1. **Check Strategy & Roadmap:** Review `docs/07-roadmap/MILESTONES.md` to identify the active milestone.
2. **Verify/Author RFC:** Ensure a feature specification exists under `docs/05-rfc/` using `0000-rfc-template.md`.
3. **Verify/Author ADR:** If making major technical or infrastructure choices, record an ADR under `docs/04-adr/` using `0000-adr-template.md`.
4. **Verify/Author Epic:** Ensure actionable tasks are listed under `docs/08-epics/`.
5. **Schema-First Contracts:** Define Protobuf, AsyncAPI, or OpenAPI schemas in `api/contracts/` before microservice implementation.
6. **Implement & Update Checklist:** After writing tests and code, mark completed tasks as `[x]` in the corresponding Epic document under `docs/08-epics/`.

---

## 2. Technical Principles & Standards (from PROJECT_BIBLE.md)

* **Project Bible First:** Always align with `docs/00-governance/PROJECT_BIBLE.md`.
* **Schema-First Mandate:** Never write microservice endpoints or event consumers without schema definitions in `api/contracts/`.
* **Event Naming Convention:** Must strictly follow `<domain>.<entity>.<past_tense_action>` (e.g., `production.work_order.completed`, `quality.ncr.raised`).
* **Entity Primary Keys:** Use UUIDv7 (time-ordered UUIDs) for all entity primary keys.
* **Edge Resilience:** Code in `platform/edge-runtime` must operate seamlessly offline using local store-and-forward SQLite buffers.
* **No Swallowing Errors:** Log explicit tracebacks and status codes. Never use silent fallbacks or ignore failing assertions.

---

## 3. Documentation Governance & File Naming

* Keep folder numbers sequential: `00-governance`, `01-overview`, `02-domain`, `03-architecture`, `04-adr`, `05-rfc`, `06-api`, `07-roadmap`, `08-epics`, `09-developer-guide`.
* Use semantic uppercase names for core entrypoints (`PROJECT_BIBLE.md`, `PRODUCT_VISION.md`).
* Use serial IDs for RFCs (`0001-xxxx.md`), ADRs (`0001-xxxx.md`), and Epics (`EPIC-001-xxxx.md`).
