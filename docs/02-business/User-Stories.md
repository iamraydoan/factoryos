# Business Domain: User Stories

## 1. Overview
Captures real user needs using the standard format: **"As a [Role], I want to [Action] so that [Value]."**

### Terminal & Device Decisions
* **Operator terminals:** Fixed industrial touchscreen monitors (15"–21") mounted at each Station. These are ruggedized, always-on displays — not personal phones.
* **Supervisor & Manager:** Desktop workstation browser + mobile-responsive dashboard on company tablet/phone for floor walkarounds.
* **Maintenance Technician & Warehouse Operator:** Handheld Android device (rugged smartphone or tablet) for mobile-first workflows.
* **Offline mode:** Required for Operator terminals and Maintenance/Warehouse handhelds. Network loss must not halt production execution. The system must buffer locally and sync when connectivity is restored.
* **Language:** English only for the initial release (v1.0). Multi-language support planned for v2.0.

---

## 2. Production Execution

| ID | User Story | Priority |
|---|---|---|
| US-001 | As an **Operator**, I want to see step-by-step work instructions on a screen at my station so that I do not need to reference paper travelers. | High |
| US-002 | As an **Operator**, I want to scan a QR code on materials to verify I am using the correct components before assembly. | High |
| US-003 | As an **Operator**, I want the terminal to work offline so that a network outage does not stop my production task. | High |
| US-004 | As a **Shift Leader**, I want to view the live status of all Stations on my Line so that I can resolve issues without waiting for escalation. | High |
| US-005 | As a **Shift Leader**, I want to generate a shift handover summary so that the incoming shift knows exactly where each WO stands. | Medium |
| US-006 | As a **Production Supervisor**, I want to view the real-time progress of every active Work Order on the floor so that I can identify delays before they escalate. | High |
| US-007 | As a **Production Supervisor**, I want to receive an alert when a Station has been stopped for longer than a defined threshold so that I can dispatch support immediately. | High |
| US-008 | As a **Production Supervisor**, I want to approve an Emergency Work Order and have it automatically inserted at the top of the active queue. | High |
| US-009 | As a **Production Planner**, I want to view a finite capacity schedule across all Lines so that I can resolve conflicts before they cause delays. | High |
| US-010 | As a **Production Planner**, I want to receive an automatic alert when a Work Order is at risk of missing its due date so that I can adjust the schedule proactively. | High |

---

## 3. Quality

| ID | User Story | Priority |
|---|---|---|
| US-020 | As a **Quality Inspector**, I want to complete an inspection checklist on the station terminal so that I can eliminate paper-based recording. | High |
| US-021 | As a **Quality Inspector**, I want to raise an NCR in one action when a defect is detected so that the affected lot is automatically quarantined. | High |
| US-022 | As a **Quality Inspector**, I want to select a disposition (Rework / Scrap / Use-As-Is) on the NCR so that the unit is routed correctly. | High |
| US-023 | As a **Quality Manager**, I want to approve Use-As-Is dispositions so that no non-conforming unit ships without explicit sign-off. | High |
| US-024 | As a **Quality Manager**, I want to look up the full component genealogy of any serial unit or lot instantly so that I can respond to a customer recall within minutes. | High |
| US-025 | As a **Quality Manager**, I want to view SPC trend charts for critical process parameters so that I can detect drift before defects occur. | Medium |

---

## 4. Maintenance

| ID | User Story | Priority |
|---|---|---|
| US-030 | As a **Maintenance Technician**, I want to receive a push notification on my handheld device when a machine fault is raised so that I can respond within the SLA. | High |
| US-031 | As a **Maintenance Technician**, I want to view the full repair history of an asset before starting work so that I can understand recurring failures. | Medium |
| US-032 | As a **Maintenance Technician**, I want to complete a digital PM checklist on my handheld so that I can confirm each step is done without paper. | Medium |
| US-033 | As a **Maintenance Manager**, I want to schedule PM tasks based on runtime hours so that I can proactively avoid unplanned downtime. | Medium |
| US-034 | As a **Maintenance Manager**, I want to track open Work Orders by severity so that Critical faults are resolved within the 15-minute SLA. | High |

---

## 5. Warehouse & Material

| ID | User Story | Priority |
|---|---|---|
| US-040 | As a **Warehouse Operator**, I want to scan incoming material to receive a lot into the system so that inventory is automatically updated. | High |
| US-041 | As a **Warehouse Operator**, I want to print a QR label for each received lot so that it can be scanned throughout the production floor. | High |
| US-042 | As a **Warehouse Operator**, I want to stage material to a specific Line buffer and confirm via scan so that Operators can verify the correct material at the Station. | High |

---

## 6. Analytics & Reporting

| ID | User Story | Priority |
|---|---|---|
| US-050 | As a **Plant Manager**, I want to see the real-time OEE for each production line so that I can understand overall plant efficiency at a glance. | High |
| US-051 | As a **Plant Manager**, I want to compare OEE across shifts, lines, and products so that I can identify improvement opportunities. | Medium |
| US-052 | As a **Plant Manager**, I want to receive an automated daily performance summary so that I do not need to manually compile reports. | Low |
