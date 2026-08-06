# Strategic Engineering Roadmap (Years 1–3)

## Milestone Breakdown Matrix

```
   MILESTONE      HORIZON                      PRIMARY FOCUS                                CORE ROI
-------------- -------------- ---------------------------------------------- ---------------------------------------
  Milestone 1   Months 0 - 6   Telemetry Ingestion & Real-Time OEE Observability Instant machine utilization visibility
  Milestone 2   Months 6 - 12  Guided Execution, Operator UI & Lot Tracking   Zero paper travelers, ground-truth data
  Milestone 3   Months 12 - 18 Reactive Workflows, Quality NCR & CMMS          Sub-second inter-departmental auto-sync
  Milestone 4   Months 18 - 24 Finite Capacity Dispatching & Re-Routing        Dynamic line balancing & bottleneck fix
  Milestone 5   Months 24 - 36 Multi-Plant Federation & Open Platform SDK      Global enterprise control tower
```

---

## Detailed Milestone Objectives

### Milestone 1: Telemetry Core (Months 0–6)
- **Goal:** Non-invasive machine observability.
- **Key Modules:** `edge-runtime`, `analytics-engine`, telemetry ingestion pipeline.
- **Success Criteria:** 100 continuous machine streams ingested with zero message loss over 7 days.

### Milestone 2: Guided Shop-Floor Execution (Months 6–12)
- **Goal:** Replace paper travelers and establish digital operator terminals.
- **Key Modules:** `production-service`, operator workstation UI, lot tracking.
- **Success Criteria:** Execute a full 5-station work order run digitally without paper sign-offs.

### Milestone 3: Reactive Departmental Orchestration (Months 12–18)
- **Goal:** Automate departmental handoffs via durable workflows.
- **Key Modules:** `workflow-engine`, `quality-service`, `maintenance-service`, `warehouse-service`.
- **Success Criteria:** Trigger quality non-conformance quarantine automatically within 100ms of metric breach.
