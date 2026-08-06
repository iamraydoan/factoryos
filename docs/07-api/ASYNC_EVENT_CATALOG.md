# Async Event Catalog & Schema Standards

## Domain Event Payload Specifications

All events published on the FactoryOS event platform adhere to standard CloudEvents spec wrappers.

---

### Core Domain Event Examples

#### 1. `factoryos.production.work_order.state_changed.v1`
* **Producer:** `production-service`
* **Description:** Emitted whenever a work order transitions state (e.g., Created -> Released -> In-Progress -> Paused -> Completed).

#### 2. `factoryos.telemetry.machine_metric.ingested.v1`
* **Producer:** `edge-runtime`
* **Description:** High-frequency machine sensor readings (temperatures, spindle speeds, cycle times, vibration data).

#### 3. `factoryos.quality.inspection.failed.v1`
* **Producer:** `quality-service`
* **Description:** Emitted when a station or line quality check fails tolerance boundaries.

#### 4. `factoryos.maintenance.work_order.created.v1`
* **Producer:** `maintenance-service` / `workflow-engine`
* **Description:** Emitted when a corrective or preventive maintenance ticket is opened.

#### 5. `factoryos.warehouse.kanban.pull_signaled.v1`
* **Producer:** `warehouse-service` / `production-service`
* **Description:** Emitted when a station inventory level drops below safety threshold, requesting raw material staging.
