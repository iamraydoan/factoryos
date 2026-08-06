# FactoryOS Strategic Milestones & Evolution Roadmap

> Each Milestone delivers a complete, end-to-end vertical slice of value. Domain modules referenced below map to `docs/03-domain/`.

---

## 🎯 Milestone 1: Resource Foundation & Real-Time Telemetry (Months 0–6)

* **Focus:** Establish the ISA-95 Resource Management layer (Equipment Hierarchy, Personnel, Material Definition), deploy the Edge Runtime for OPC-UA/MQTT telemetry ingestion, and deliver real-time OEE computation per Work Unit.
* **Domain Modules:** `01-resource-management/`, `platform/edge-runtime`, `analytics-engine`
* **Primary RFCs:** `RFC-0001` (Asset Ontology & Telemetry Ingestion)
* **Primary Epics:** `EPIC-001` (Platform Foundation), `EPIC-002` (Resource Management & Telemetry)
* **Key Deliverable:** Real-time OEE dashboard per Work Unit / Work Center. Plant Manager can see live Availability, Performance, and Quality metrics without leaving their desk.

---

## 🎯 Milestone 2: Production Execution & Inventory Traceability (Months 6–12)

* **Focus:** Full Production Operations lifecycle (Production Request → Schedule → Work Order → Operation → Execution), Operator Terminal UI, and Inventory Operations (Material Lot staging, consumption, and As-Built Genealogy).
* **Domain Modules:** `02-production-operations/`, `05-inventory-operations/`
* **Primary RFCs:** `RFC-0002` (Work Order Execution Engine), `RFC-0003` (Material Genealogy Tracking)
* **Primary Epics:** `EPIC-003` (Production Operations), `EPIC-004` (Inventory & Genealogy)
* **Key Deliverable:** An Operator can execute a complete 5-step Work Order digitally, scan material QR codes, and a Quality Manager can run a full lot recall query returning all affected serial units in < 2 seconds.

---

## 🎯 Milestone 3: Closed-Loop Quality & Maintenance (Months 12–18)

* **Focus:** Full Quality Operations (Test Specification → Inspection → NCR → Disposition → CAPA) and Maintenance Operations (Physical Asset registry, Maintenance Request SLA dispatch, PM Scheduling, Downtime tracking).
* **Domain Modules:** `03-quality-operations/`, `04-maintenance-operations/`
* **Primary RFCs:** `RFC-0004` (Quality Operations), `RFC-0005` (Maintenance Operations)
* **Primary Epics:** `EPIC-005` (Quality Operations), `EPIC-006` (Maintenance Operations)
* **Key Deliverable:** NCR raised automatically triggers lot quarantine and Maintenance WO dispatch within 15 seconds. PM schedules run automatically based on Physical Asset runtime hours.

---

## 🎯 Milestone 4: Finite Capacity Scheduling & Edge Resilience (Months 18–24)

* **Focus:** Advanced Production Scheduling with finite capacity optimization, dynamic Work Unit re-routing on bottleneck detection, and full offline edge execution (local buffer, no cloud dependency).
* **Domain Modules:** `02-production-operations/Production-Schedule.md` (advanced), `platform/edge-runtime` (offline mode)
* **Primary RFCs:** `RFC-0006` (Finite Capacity Scheduling)
* **Primary Epics:** `EPIC-007` (Finite Scheduling & Edge Resilience)
* **Key Deliverable:** Plant continues producing at full rate through a 4-hour cloud connectivity outage, with zero data loss and automatic sync on reconnect.

---

## 🎯 Milestone 5: Multi-Plant Control Tower & Industrial AI (Months 24–36)

* **Focus:** Multi-site aggregation, cross-plant OEE benchmarking, predictive maintenance AI (condition-based PM), open developer plugin ecosystem.
* **Domain Modules:** All (multi-site extension), `04-maintenance-operations/PM-Schedule.md` (condition-based trigger)
* **Primary RFCs:** `RFC-0007` (Multi-Plant Federation), `RFC-0008` (Industrial AI & Predictive Maintenance)
* **Primary Epics:** `EPIC-008` (Multi-Plant Control Tower)
* **Key Deliverable:** Plant Manager can view aggregated OEE, NCR rates, and MTBF across all sites from a single dashboard. AI model predicts machine failure 48 hours in advance.
