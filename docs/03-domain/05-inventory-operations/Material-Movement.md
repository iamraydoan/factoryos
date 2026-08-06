# Inventory Operations — Material Movement

> **ISA-95 Reference:** Part 3, §8 — Inventory Operations (Material Movement)

## 1. Overview
**Material Movement** tracks every physical transfer of material within the facility: from goods receipt to staging at a Work Unit, and from a staging bin into production consumption. Every movement creates an auditable record.

---

## 2. Movement Types

| Movement Type | From | To | Trigger |
|---|---|---|---|
| **Goods Receipt** | External / Supplier | Receiving Dock / Warehouse | PO delivery |
| **Put-Away** | Receiving Dock | Storage Location | Receiving inspection pass |
| **Staging** | Storage Location | Work Unit Staging Bin | Production WO material requirement |
| **Consumption** | Work Unit Staging Bin | Production (consumed) | Operator QR scan at Operation start |
| **Return to Stock** | Work Unit Staging Bin | Storage Location | Unused material from cancelled WO |
| **Transfer** | Storage Location A | Storage Location B | Internal inventory rebalancing |
| **Scrap Write-off** | Anywhere | — (written off) | NCR Scrap disposition |

---

## 3. Entity Definitions

| Entity | Description |
|---|---|
| **MaterialMovement** | A record of a quantity of a MaterialLot being transferred between two locations |
| **StorageLocation** | A named physical location in the facility (Rack A1, Staging Bin Line-3-OP-020) |
| **StagingBin** | A buffer location immediately adjacent to a Work Unit |

---

## 4. Key Events

| Event | Description |
|---|---|
| `inventory.movement.staged` | Material Lot quantity moved to a Work Unit Staging Bin |
| `inventory.movement.consumed` | Operator confirms material consumption at Operation start |
| `inventory.movement.returned_to_stock` | Unused staged material returned to warehouse |
| `inventory.movement.transferred` | Internal location-to-location transfer recorded |
