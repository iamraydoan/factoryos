# Maintenance Operations — Physical Asset

> **ISA-95 Reference:** Part 2, Section 8 — Physical Asset Object Model

## 1. Overview
A **Physical Asset** is the actual physical piece of machinery with a serial number, installation date, manufacturer warranty, and full maintenance history. This is **distinct from Equipment** (the logical scheduling unit in Resource Management) — one Physical Asset is installed *at* a Work Unit but is not the same concept.

---

## 2. Entity Definitions

| Entity | ISA-95 Term | Description |
|---|---|---|
| **PhysicalAssetClass** | Physical Asset Class | A type/model of machine (e.g., "Fanuc Robot R-2000iC/165F") |
| **PhysicalAsset** | Physical Asset | A specific machine instance: serial number, install date, location, current Work Unit |
| **AssetSpecification** | — | Technical specifications: rated capacity, power requirements, tolerance specs |
| **AssetHealthMetric** | — | Real-time sensor readings: vibration (mm/s), temperature (°C), cycle count |
| **AssetDocument** | — | Attached manuals, P&IDs, electrical schematics |

---

## 3. Physical Asset vs. Equipment (ISA-95 Distinction)

| Concept | ISA-95 Layer | Purpose |
|---|---|---|
| **Work Unit** (Equipment) | Scheduling / Routing | Capacity unit used by Production Planning |
| **Physical Asset** | Maintenance | The real machine tracked for maintenance, health, and warranty |

A Physical Asset is **installed at** a Work Unit. If the machine is replaced, the Work Unit remains — only the Physical Asset assignment changes.

---

## 4. Asset States

```text
[ ACTIVE ] ──▶ [ FAULTED ] ──▶ [ UNDER_MAINTENANCE ] ──▶ [ ACTIVE ]
                                         │
                                         ▼
                                   [ DECOMMISSIONED ]
```

---

## 5. Key Events

| Event | Description |
|---|---|
| `maintenance.asset.installed` | A Physical Asset is commissioned at a Work Unit |
| `maintenance.asset.telemetry_emitted` | Real-time sensor reading published to Historian |
| `maintenance.asset.health_alert_triggered` | A sensor reading exceeds the configured threshold |
| `maintenance.asset.decommissioned` | Asset retired from active use |
