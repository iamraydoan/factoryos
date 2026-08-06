# Bounded Context: ISA-95 Equipment Hierarchy & Asset Ontology

## 1. Overview
Defines the physical and operational hierarchy of the manufacturing facility according to ISA-95 standards.

---

## 2. Model Structure

```text
Enterprise
└── Site (Factory)
    └── Area
        └── Production Line (Work Center)
            └── Work Cell / Station
                └── Equipment Unit (Machine, PLC, Robot, Sensor)
```

---

## 3. Core Entities & Aggregates

* **Site:** Physical plant location.
* **Area:** Logical division of a site (e.g., Stamping Area, Assembly Area).
* **Line:** Production line executing specific product routings.
* **Station:** Operator workspace or automated cell on a line.
* **Asset:** Machinery, sensors, or tools associated with a station.

---

## 4. Key Events

* `factory.site.provisioned`
* `factory.line.status_changed`
* `factory.station.state_transitioned`
* `factory.asset.telemetry_emitted`
