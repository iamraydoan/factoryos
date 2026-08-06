# EPIC-005: Quality Operations

* **Milestone:** Milestone 3
* **Status:** Planned
* **Domain:** `docs/03-domain/03-quality-operations/`
* **RFCs:** `RFC-0004` (TBD)

---

## Task Checklist

### Test Specification & Inspection
- [ ] Implement TestSpecification and TestCriteria definition API (Inline, End-of-Line, Receiving).
- [ ] Implement Inspection runtime instance generation.
- [ ] Implement SPC data point streaming to Historian for numeric criteria.
- [ ] Implement Inspection UI for Operator Terminal / Quality Tablet.

### Non-Conformance (NCR)
- [ ] Implement NCR generation logic (auto-triggered on failed Inspection).
- [ ] Implement Defect classification system.
- [ ] Implement automatic MaterialLot quarantine trigger on NCR creation.
- [ ] Build Quality Manager NCR dashboard.

### Disposition & CAPA
- [ ] Implement Disposition workflow (Rework, Scrap, Use-As-Is, Return to Supplier).
- [ ] Implement automatic Work Order hold / Rework Station routing based on Disposition.
- [ ] Implement CAPA workflow state machine (Open → Root Cause → Actions Assigned → Effectiveness Review → Closed).
- [ ] Implement Use-As-Is approval matrix (requires Quality Manager sign-off).
