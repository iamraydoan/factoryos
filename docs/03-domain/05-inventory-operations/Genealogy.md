# Inventory Operations — Genealogy (As-Built Traceability)

> **ISA-95 Reference:** Part 2, §7 — Material Lot Genealogy

## 1. Overview
**Genealogy** records the full bidirectional parent-child relationship between raw material lots and the finished/semi-finished serial units they became part of. This enables instant regulatory recall response and quality root-cause traceability.

---

## 2. Genealogy Model (As-Built Tree)

```text
SerialUnit: SN-2024-00421 (Engine Block Assembly)
└── consumed MaterialLot: CRANK-LOT-0082 (Crankshaft Sub-Assembly)
│       └── consumed MaterialLot: STEEL-LOT-A1934 (Raw Steel Bar)
├── consumed MaterialLot: GASKET-LOT-0037 (Gasket Set)
└── consumed MaterialLot: BEARING-LOT-2201 (Main Bearing)
        └── consumed MaterialLot: STEEL-LOT-B0772 (Bearing Steel)
```

---

## 3. Entity Definitions

| Entity | Description |
|---|---|
| **SerialUnit** | A uniquely identified finished or semi-finished product instance |
| **GenealogyNode** | A node in the As-Built tree: a (parent SerialUnit, child MaterialLot/SerialUnit, quantity consumed) record |
| **GenealogyTree** | The complete directed acyclic graph (DAG) for a given SerialUnit, traversable in both directions |

---

## 4. Traceability Queries

| Query Direction | Question Answered |
|---|---|
| **Forward (Where-Used)** | "Which finished serial units contain MaterialLot X?" → Used for recall |
| **Backward (As-Built)** | "What raw materials were used to build SerialUnit Y?" → Used for root cause |

Both queries must return results in **< 2 seconds** regardless of tree depth (target SLA for recall response).

---

## 5. Genealogy Record Triggers
A genealogy link is written when:
- An Operator confirms material consumption at an Operation (QR scan).
- A Work Order is completed and the finished serial unit is created.

---

## 6. Key Events

| Event | Description |
|---|---|
| `inventory.genealogy.component_consumed` | A MaterialLot is linked as a child component of a SerialUnit during production |
| `inventory.genealogy.serial_unit_created` | A new SerialUnit is born at Work Order completion |
| `inventory.genealogy.recall_query_executed` | A recall forward-trace query is run; result set logged for audit |
