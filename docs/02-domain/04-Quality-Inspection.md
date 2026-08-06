# Bounded Context: Quality Inspection & Non-Conformance (NCR)

## 1. Overview
Manages inline quality checklists, automated vision inspection triggers, Statistical Process Control (SPC), and Non-Conformance Reports (NCR) with CAPA workflows.

---

## 2. Core Entities & Aggregates

* **QualityChecklist:** Standardized inspection criteria for an operation step.
* **InspectionResult:** Pass/Fail outcome with attached telemetry or imagery.
* **NonConformanceReport (NCR):** Ticket raised when quality specs are breached.
* **CAPA:** Corrective and Preventive Action workflow.

---

## 3. Key Events

* `quality.inspection.passed`
* `quality.inspection.failed`
* `quality.ncr.raised`
* `quality.capa.assigned`
