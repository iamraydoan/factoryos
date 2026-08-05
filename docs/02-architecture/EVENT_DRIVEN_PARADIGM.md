# Event-Driven Architecture Paradigm

## 1. Why Event-Driven?
Traditional request-response architectures fail on manufacturing shop floors due to high-frequency telemetry data streams, unreliable physical networks, and complex multi-departmental side effects.

An Event-Driven Architecture (EDA) allows FactoryOS to:
* React to machine telemetry anomalies in sub-second timelines.
* Fan out events across multiple downstream domain consumers (Analytics, CMMS, Quality, Staging) without hard service couplings.
* Maintain complete chronological history of physical state changes for regulatory audit trails.

---

## 2. Event Classification Taxonomy

```
                              ┌──────────────────┐
                              │  Factory Event   │
                              └────────┬─────────┘
                                       │
            ┌──────────────────────────┼──────────────────────────┐
            ▼                          ▼                          ▼
  +------------------+       +------------------+       +------------------+
  | Telemetry Metric |       |  Domain State    |       | Operational Command|
  | (High-Frequency) |       | (Business Event) |       | (Intended Action) |
  +------------------+       +------------------+       +------------------+
  E.g. Spindle Temp,          E.g. WorkOrderStarted,     E.g. HaltStationCommand,
  Machine RPM, Current       LotQuarantined, Faulted    DispatchAGVCommand
```

### A. Telemetry Metric Events (High-Volume, Time-Series)
* **Characteristics:** High throughput (1,000 to 100,000 msgs/sec), raw sensor inputs, ephemeral unless aggregated.
* **Storage Path:** Edge Agent -> Streaming Event Bus -> Time-Series Metrics Database.

### B. Domain State Events (Business State Mutations)
* **Characteristics:** Strict schema validation, transactional consistency, long-term retention.
* **Examples:** `WorkOrderReleased`, `MachineStateChanged`, `LotQuarantined`, `QualityInspectionFailed`.
* **Storage Path:** Domain Service -> Event Bus -> State Store + Audit Log Ledger.

### C. Operational Command Events (Action Execution Requests)
* **Characteristics:** Target-specific, requires acknowledgment, subject to timeout and compensation workflows.
* **Examples:** `HaltStationLine`, `TriggerMaterialReplenishment`, `DispatchTechnician`.
* **Storage Path:** Durable Workflow Engine -> Target Queue / Edge Gateway.
