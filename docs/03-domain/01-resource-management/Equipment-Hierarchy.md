# Resource Management — Equipment Hierarchy

> **ISA-95 Reference:** Part 2, Section 5 — Equipment Object Model

## 1. Overview
Defines the logical operational hierarchy of the manufacturing facility. Equipment entities represent **schedulable capacity units** used for planning and routing — distinct from Physical Assets (see `04-maintenance-operations/Physical-Asset.md`).

---

## 2. ISA-95 Equipment Hierarchy

```text
Enterprise
└── Site                    (physical plant / facility)
    └── Area                (functional zone: Stamping, Assembly, Painting)
        └── Work Center     (group of Work Units with similar capability)
            └── Work Unit   (individual operator workstation or automated cell)
                └── Equipment Class → links to Physical Asset
```

---

## 3. Entity Definitions

| Entity | ISA-95 Term | Description |
|---|---|---|
| **Enterprise** | Enterprise | Top-level organizational owner |
| **Site** | Site | A physical manufacturing plant location |
| **Area** | Area | A functional zone within a Site (e.g., Machining Area) |
| **WorkCenter** | Work Center | A schedulable group of Work Units with shared capability |
| **WorkUnit** | Work Unit | The finest schedulable production location; maps to a physical Station |
| **EquipmentClass** | Equipment Class | Capability definition (e.g., "CNC Lathe ≥ 5-axis") used to find qualified Work Units |

---

## 4. Work Unit States

```text
[ AVAILABLE ] ──▶ [ ALLOCATED ]  (Work Order assigned)
     ▲                  │
     │                  ▼
[ RELEASED ] ◀── [ IN_PRODUCTION ]
                        │
                        ▼
                  [ FAULTED ]  ──▶ triggers Maintenance Request
```

---

## 5. Key Events

| Event | Description |
|---|---|
| `resource.site.provisioned` | A new Site is registered in the system |
| `resource.work_center.created` | A Work Center is defined under an Area |
| `resource.work_unit.status_changed` | Work Unit transitions state (e.g., Available → Faulted) |
| `resource.equipment_class.updated` | Capability specification updated for a Work Center |
