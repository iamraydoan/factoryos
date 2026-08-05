# ADR 0003: Adoption of Durable Execution Engine

* **Status:** Accepted
* **Deciders:** Founding CTO, Principal Software Architect
* **Date:** 2026-08-05

## Context & Problem Statement
Physical manufacturing processes involve complex, long-running operational workflows that cross physical and digital boundaries. For example, when a station reports an out-of-spec scrap rate:
1. Halt station execution state.
2. Dispatch a quality technician.
3. Wait for inspection results (could take minutes or hours).
4. If passed -> Resume station.
5. If failed -> Quarantine batch, update BOM routing, notify line supervisor.

If microservices restart or experience network issues during step 3, state must not be lost or duplicated.

## Decision Outcome
Adopt a **Durable Execution Engine Framework** (e.g., Temporal abstraction core) in `workflow-engine`.

* Workflows maintain execution state durably across process crashes and server outages.
* Built-in retry handling, timers, signals, and saga compensation patterns eliminate custom state machine boilerplate in domain services.
