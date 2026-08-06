# Business Domain: Manufacturing Glossary

## 1. Overview
Defines the shared business vocabulary used throughout FactoryOS. All engineers, QC staff, and stakeholders must use this glossary as the single source of truth for terminology.

---

## 2. Production Terms

| Term | Definition |
|---|---|
| **Work Order (WO)** | A manufacturing directive to produce a specified quantity of a product following a defined Routing. |
| **Routing** | The ordered sequence of Operations required to manufacture a product from raw material to finished goods. |
| **Operation** | A single discrete manufacturing step within a Routing, performed at a specific Station or Work Center. |
| **Bill of Materials (BOM)** | The structured list of raw materials, components, and quantities required to produce one unit of a product. |
| **Work Center** | A logical grouping of Stations with similar capabilities, used for capacity planning and scheduling. |
| **Station** | A physical workstation on the shop floor where an Operator performs one or more Operations. |
| **Shift** | A defined working period (e.g., Day Shift, Night Shift) during which Operators and Lines are active. |
| **Traveler** | A paper or digital document that accompanies a Work Order through the production floor, recording completion of each step. |

---

## 3. Quality Terms

| Term | Definition |
|---|---|
| **Non-Conformance Report (NCR)** | A formal record raised when a product, lot, or process does not meet defined quality specifications. |
| **CAPA** | Corrective and Preventive Action — the structured investigation and remediation process triggered by an NCR. |
| **Inspection Checklist** | A structured list of quality criteria to be evaluated at a specific Operation or Station. |
| **Lot** | A group of products manufactured under the same conditions from the same batch of materials; traceable as a unit. |
| **Serial Number** | A unique identifier assigned to an individual unit for item-level traceability. |
| **As-Built Record** | The actual record of which material lots and components were consumed to produce a specific serialized unit. |
| **SPC** | Statistical Process Control — monitoring of process parameters over time using control charts to detect drift before defects occur. |
| **Recall** | A quality event requiring the identification and retrieval of all affected lots or serial units from the field or production. |

---

## 4. Maintenance Terms

| Term | Definition |
|---|---|
| **Preventive Maintenance (PM)** | Scheduled maintenance tasks performed proactively to prevent equipment failure. |
| **Corrective Maintenance (CM)** | Maintenance performed reactively after an equipment failure has occurred. |
| **MTBF** | Mean Time Between Failures — the average operating time between two consecutive equipment failures. |
| **MTTR** | Mean Time To Repair — the average time required to restore equipment to operational status after a failure. |
| **Work Order (Maintenance)** | A directive to perform a specific maintenance task on a defined asset. |
| **Downtime** | Any period during which a machine or Station is not producing due to unplanned stoppage or maintenance. |

---

## 5. Asset & Telemetry Terms

| Term | Definition |
|---|---|
| **OEE** | Overall Equipment Effectiveness = Availability × Performance × Quality. The primary KPI for manufacturing efficiency. |
| **Availability** | The percentage of planned production time during which the equipment is actually running. |
| **Performance** | The ratio of actual production speed to the theoretical maximum speed. |
| **Quality (OEE)** | The ratio of good units produced to total units started (excluding scrap and rework). |
| **OPC-UA** | Open Platform Communications Unified Architecture — the industrial standard protocol for communicating with PLCs and machines. |
| **MQTT** | A lightweight publish-subscribe messaging protocol used for IoT sensor telemetry. |
| **PLC** | Programmable Logic Controller — the embedded computer that controls automated machinery on the shop floor. |
| **Historian** | A time-series database system that stores and queries high-frequency machine telemetry data. |
| **Telemetry** | Continuous streams of real-time sensor readings emitted by machines (e.g., temperature, vibration, cycle count). |
