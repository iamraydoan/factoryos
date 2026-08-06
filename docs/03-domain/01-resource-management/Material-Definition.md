# Resource Management — Material Definition

> **ISA-95 Reference:** Part 2, Section 7 — Material Object Model

## 1. Overview
Defines the **master data** for all materials and products: their classification, specifications, and Bill of Materials structure. This is the static reference layer consumed by Inventory Operations (lot tracking) and Production Operations (BOM verification at execution time).

---

## 2. Entity Definitions

| Entity | ISA-95 Term | Description |
|---|---|---|
| **MaterialClass** | Material Class | Category of material (e.g., Raw Material, Sub-Assembly, Finished Good) |
| **MaterialDefinition** | Material Definition | A specific material or product with its specification and unit of measure |
| **BillOfMaterials (BOM)** | — | Hierarchical list of MaterialDefinitions required to produce one unit of a parent MaterialDefinition |
| **BOMComponent** | — | A line item in a BOM specifying a child MaterialDefinition and required quantity |
| **ProductRoutingSpec** | Process Segment Specification | Links a MaterialDefinition to the sequence of Work Centers required to produce it |

---

## 3. Hierarchy Example

```text
MaterialDefinition: "Engine Block Assembly" (Finished Good)
└── BOM
    ├── BOMComponent: "Cylinder Block" × 1     (Sub-Assembly)
    ├── BOMComponent: "Piston Set"    × 4     (Raw Material)
    └── BOMComponent: "Gasket Kit"    × 1     (Consumable)
```

---

## 4. Key Events

| Event | Description |
|---|---|
| `resource.material_definition.created` | A new product or material is registered |
| `resource.bom.revised` | A BOM version is updated (triggers version control) |
| `resource.product_routing.published` | A routing spec is approved and made active |
