# EPIC-002: Resource Management & Real-Time Telemetry

* **Milestone:** Milestone 1
* **Status:** Planned
* **Domain:** `docs/03-domain/01-resource-management/`
* **RFC:** `RFC-0001`

---

## Task Checklist

### Equipment Hierarchy (ISA-95 Resource Model)
- [ ] Implement CRUD API for Site, Area, Work Center, Work Unit entities.
- [ ] Implement Equipment Class definition and Work Unit capability assignment.
- [ ] Implement Work Unit status state machine: Available → Allocated → In Production → Faulted.
- [ ] Link Work Unit to Physical Asset (installation record).

### Personnel & Qualification
- [ ] Implement Person and PersonClass (role) management.
- [ ] Implement QualificationRecord: certify a Person for a PersonClass at a Work Center.
- [ ] Implement QualificationRecord expiry tracking and alert.
- [ ] Implement Shift and ShiftAssignment management.

### Material Definition
- [ ] Implement MaterialClass and MaterialDefinition CRUD API.
- [ ] Implement Bill of Materials (BOM) with versioning support.
- [ ] Implement ProductRoutingSpec: link MaterialDefinition to Work Center sequence.

### Telemetry & OEE
- [ ] Build Go Edge Runtime OPC-UA collector (`platform/edge-runtime`).
- [ ] Configure SQLite store-and-forward buffer for offline tolerance.
- [ ] Implement MQTT ingestion endpoint for sensor telemetry.
- [ ] Implement TimescaleDB hypertable for high-frequency telemetry storage.
- [ ] Implement real-time OEE streaming aggregator (Availability × Performance × Quality) per Work Unit.
- [ ] Implement OEE alert when Work Unit drops below configured threshold.
