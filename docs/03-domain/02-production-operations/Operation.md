# Production Operations — Operation

> **ISA-95 Reference:** Part 2, Section 4 — Job Element (within a Job Order)

## 1. Overview
An **Operation** (ISA-95: Job Element) is a single, discrete manufacturing step within a Work Order. It is performed at a specific Work Unit, by a specific Operator, following a defined work instruction. Operations are the finest unit of execution tracked by FactoryOS.

---

## 2. Entity Definitions

| Entity | ISA-95 Term | Description |
|---|---|---|
| **Operation** | Job Element | A single step in the Routing, linked to a Work Unit and PersonClass |
| **WorkInstruction** | — | The digital step-by-step guide displayed to the Operator at the terminal |
| **OperationBOMConsumption** | — | The specific material lots consumed during this Operation |

---

## 3. Operation State Machine

```text
[ PENDING ] ──▶ [ ACTIVE ] ──▶ [ COMPLETED ]
                    │
                    ▼
               [ HELD ]  (triggered by Quality hold or material shortage)
                    │
                    ▼
               [ RESUMED ]
```

---

## 4. Routing Sequence
Operations within a Work Order execute **sequentially by default**. Parallel execution is supported where the Routing spec defines no dependency between steps.

```text
WO-001
├── OP-010: Machining       (Work Center A)
├── OP-020: Deburring       (Work Center A)
├── OP-030: Sub-Assembly    (Work Center B)
└── OP-040: Final Inspection (Quality Station)  ← triggers Quality Operation
```

---

## 5. Key Events

| Event | Description |
|---|---|
| `production.operation.started` | Operator confirms start on terminal |
| `production.operation.completed` | Operator submits completion; output quantity recorded |
| `production.operation.held` | Operation paused due to quality, material, or machine issue |
| `production.operation.skipped` | Operation bypassed with Supervisor authorization |
