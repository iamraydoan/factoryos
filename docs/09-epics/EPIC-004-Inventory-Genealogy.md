# EPIC-004: Inventory Operations & Genealogy Tracking

* **Milestone:** Milestone 2
* **Status:** Planned
* **Domain:** `docs/03-domain/05-inventory-operations/`
* **RFCs:** `RFC-0003`

---

## Task Checklist

### Material Lot Management
- [ ] Implement MaterialLot CRUD with auto-generated UUIDv7 `lot_id`.
- [ ] Implement Lot status state machine: Received → Available / Quarantined → Staged → Consumed / Scrapped.
- [ ] Implement Lot expiration tracking and automated quarantine trigger.

### Material Movement
- [ ] Implement StorageLocation and StagingBin configuration.
- [ ] Implement MaterialMovement API to track transfers (Goods Receipt, Put-Away, Staging, Return-to-Stock).
- [ ] Build Warehouse Operator mobile UI for scan-and-receive and scan-and-stage workflows.

### Genealogy & Traceability
- [ ] Implement As-Built Genealogy DAG (Directed Acyclic Graph) linking API.
- [ ] Auto-link `MaterialLot` to `SerialUnit` upon component consumption during an Operation.
- [ ] Auto-create parent `SerialUnit` upon Work Order completion.
- [ ] Implement high-performance Forward (Where-Used) recall query (< 2 sec SLA).
- [ ] Implement high-performance Backward (As-Built) root cause query (< 2 sec SLA).
