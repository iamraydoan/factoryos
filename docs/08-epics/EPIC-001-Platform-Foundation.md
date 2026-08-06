# EPIC-001: Platform Architecture & Infrastructure Foundation

* **Milestone:** Milestone 1
* **Status:** In Progress

---

## Task Checklist

### Core Setup & Governance
- [x] Create `PROJECT_BIBLE.md` and repository standards.
- [x] Define Protobuf / AsyncAPI contracts directory hierarchy.
- [ ] Configure PostgreSQL with Transactional Outbox table schema.
- [ ] Setup Kafka / Redpanda event streaming infrastructure.

### Platform Services
- [ ] Implement `platform-sdk` event producer/consumer wrappers.
- [ ] Implement durable state machine workflow executor in `platform/workflow-engine`.
