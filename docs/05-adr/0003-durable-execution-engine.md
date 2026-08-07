# ADR-0003: Adoption of Durable Execution Engine

* **Status:** Accepted
* **Date:** 2026-08-05
* **Authors:** Founding CTO, Principal Software Architect

---

## 1. Context

Physical manufacturing processes involve complex, long-running operational workflows that cross physical and digital boundaries. For example, when a station reports an out-of-spec scrap rate:
1. Halt station execution state.
2. Dispatch a quality technician.
3. Wait for inspection results (could take minutes or hours).
4. If passed -> Resume station.
5. If failed -> Quarantine batch, update BOM routing, notify line supervisor.

If microservices restart or experience network issues during step 3, state must not be lost or duplicated.

---

## 2. Decision

Adopt a **Durable Execution Engine Framework** (e.g., Temporal abstraction core) in `workflow-engine`.

* Workflows maintain execution state durably across process crashes and server outages.
* Built-in retry handling, timers, signals, and saga compensation patterns eliminate custom state machine boilerplate in domain services.

---

## 3. Consequences

### Positive Impacts
- Guaranteed workflow execution and state resilience across server reboots.
- Built-in compensation (Saga pattern) for multi-service transactions.

### Negative Impacts & Trade-offs
- Operational overhead of running a Temporal cluster and worker nodes.

---

## 4. Alternatives Considered

* **Custom DB polling state machine:** Rejected due to high risk of race conditions, deadlocks, and unhandled edge-case failures.

