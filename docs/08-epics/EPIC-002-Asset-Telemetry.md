# EPIC-002: ISA-95 Asset Hierarchy & Telemetry Ingestion

* **Milestone:** Milestone 1
* **Status:** Planned

---

## Task Checklist

### Asset Domain Model
- [ ] Implement Asset Hierarchy CRUD API (Site, Area, Line, Station).
- [ ] Build Equipment Status state machine (Running, Idle, Faulted, Off).

### Telemetry & Ingestion
- [ ] Build Go Edge Runtime OPC-UA collector (`platform/edge-runtime`).
- [ ] Configure SQLite store-and-forward buffer for offline tolerance.
- [ ] Implement TimescaleDB hypertable for high-frequency telemetry storage.
- [ ] Compute real-time OEE (Availability, Performance, Quality) streaming aggregator.
