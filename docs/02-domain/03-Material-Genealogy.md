# Bounded Context: Material Staging & Lot Genealogy (As-Built Tree)

## 1. Overview
Tracks raw material inventory allocation, shop floor material staging, lot/serial number tracking, and full bidirectional genealogy (where-used tree).

---

## 2. Core Entities & Aggregates

* **MaterialLot:** Batched inventory lot with supplier certificate and expiration data.
* **SerialUnit:** Individually tracked finished or semi-finished component.
* **GenealogyTree:** Parent-child consumption mapping of raw components to final serial units.
* **StagingBin:** Physical buffer location near a work station on the floor.

---

## 3. Key Events

* `warehouse.lot.received`
* `warehouse.staging.allocated`
* `production.component.consumed`
* `genealogy.as_built.linked`
