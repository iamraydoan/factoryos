# EPIC-003: Production Operations

* **Milestone:** Milestone 2
* **Status:** In Progress
* **Domain:** `docs/03-domain/02-production-operations/`
* **RFCs:** `RFC-0002`

---

## Task Checklist

### Production Request (ERP Integration)
- [ ] Implement Production Request ingestion API (from ERP via Protobuf/REST).
- [ ] Implement Production Request validation: BOM check, material availability, capacity check.
- [ ] Implement Production Request state machine: Received → Validated → Scheduled → Fulfilled.

### Production Schedule
- [ ] Implement Production Schedule CRUD with Work Center time-slot assignment.
- [ ] Implement priority-based sequencing: customer tier A/B/C → due date.
- [ ] Implement at-risk Work Order alert when schedule slot cannot meet due date.
- [ ] Implement Emergency Work Order insertion at queue top with Plant Manager approval.

### Work Order (Job Order)
- [ ] Implement Work Order CRUD and full state machine (Draft → Released → Dispatched → In Progress → Completed → Closed).
- [ ] Implement dispatch validation: Work Unit capability match + Operator qualification check + materials staged.
- [ ] Implement Work Order hold / resume with reason recording.

### Operation (Job Element)
- [ ] Implement Operation model and sequential routing execution engine.
- [ ] Implement Work Instruction delivery to Operator terminal per Operation.
- [ ] Implement Operation state machine: Pending → Active → Completed / Held.
- [ ] Implement parallel Operation support where Routing spec defines no dependency.

### Execution (Production Response)
- [ ] Implement JobResponse recording: actual quantity good/reject per Work Order.
- [ ] Implement JobElementResponse: actual duration, operator, material consumed per Operation.
- [ ] Implement ProductionEvent ledger for timestamped floor events (start, stop, hold, fault).

### Operator Terminal UI
- [ ] Build Operator Terminal Web UI (React/Next.js) for industrial touchscreen.
- [ ] Implement QR scan material verification at Operation start.
- [ ] Implement one-tap "Report Issue" action on terminal.
- [ ] Implement offline mode: buffer execution events locally, sync on reconnect.

### Performance
- [ ] Implement OEE snapshot computation per Work Unit (streaming, every 60 seconds).
- [ ] Implement Shift Performance Summary report (auto-generated at shift end).
- [ ] Implement Work Center and Area OEE roll-up aggregation.
