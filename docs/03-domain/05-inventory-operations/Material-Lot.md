# Inventory Operations — Material Lot

> **ISA-95 Reference:** Part 2, Section 7 — Material Lot / Material Sublot

## 1. Overview
A **Material Lot** is a discrete, traceable batch of physical material received into the facility. Every unit of material that enters or moves through the production floor is tracked at the Lot (or Sublot) level, providing the foundation for the Genealogy As-Built tree.

---

## 2. Entity Definitions

| Entity | ISA-95 Term | Description |
|---|---|---|
| **MaterialLot** | Material Lot | A batch of a specific MaterialDefinition received from a supplier or produced internally |
| **MaterialSublot** | Material Sublot | A sub-division of a Lot when partial quantities are split for different destinations |
| **LotAttribute** | — | Metadata on the lot: supplier certificate number, expiry date, country of origin, inspection status |
| **LotStatus** | — | Current state of the lot: Available, Quarantined, Consumed, Scrapped |

---

## 3. Material Lot State Machine

```text
[ RECEIVED ] ──▶ [ INSPECTION_PENDING ] ──▶ [ AVAILABLE ] ──▶ [ STAGED ]
                          │                                         │
                          ▼                                         ▼
                   [ QUARANTINED ]                           [ CONSUMED ]
                          │
                 ┌─────────┴──────────┐
                 ▼                    ▼
           [ RELEASED ]          [ SCRAPPED ]
```

---

## 4. Lot Attributes

| Attribute | Required | Description |
|---|---|---|
| `lot_id` | Yes | UUIDv7 system-generated identifier |
| `material_definition_id` | Yes | Link to MaterialDefinition |
| `supplier_lot_number` | Yes | Supplier's own batch reference |
| `received_at` | Yes | Timestamp of goods receipt |
| `expiry_date` | Conditional | Required for perishable or date-sensitive materials |
| `certificate_of_conformance` | Optional | Supplier quality document reference |

---

## 5. Key Events

| Event | Description |
|---|---|
| `inventory.lot.received` | Lot scanned into the system at goods receipt |
| `inventory.lot.quarantined` | Lot placed on hold (NCR or receiving inspection failure) |
| `inventory.lot.released` | Quarantine lifted after re-inspection |
| `inventory.lot.consumed` | Lot fully consumed in production |
