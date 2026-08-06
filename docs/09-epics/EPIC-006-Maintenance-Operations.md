# EPIC-006: Maintenance Operations

* **Milestone:** Milestone 3
* **Status:** Planned
* **Domain:** `docs/03-domain/04-maintenance-operations/`
* **RFCs:** `RFC-0005` (TBD)

---

## Task Checklist

### Physical Asset & Downtime
- [ ] Implement PhysicalAsset tracking (install base, serial number, warranty).
- [ ] Implement asset swap logic (move Physical Asset between Work Units).
- [ ] Implement Downtime ledger categorized by SEMI E10 (UD, SD, ED, SB, NS).
- [ ] Implement automatic MTBF and MTTR metric calculation per asset.

### Maintenance Request & WO
- [ ] Implement auto-trigger Maintenance Request based on telemetry threshold.
- [ ] Implement auto-trigger Maintenance Request based on OEE degradation.
- [ ] Implement Maintenance Work Order (CM) dispatch to Technician.
- [ ] Build Technician mobile UI for WO acceptance, checklist, and closure.
- [ ] Implement SLA escalation logic (Critical ≤ 15m, Major ≤ 45m).

### Preventive Maintenance
- [ ] Implement PMSchedule configuration (Calendar, Runtime Hours, Cycle Count triggers).
- [ ] Implement PM trigger engine hooked to Historian / JobResponse streams.
- [ ] Implement automatic PM Work Order creation when trigger condition is met.
