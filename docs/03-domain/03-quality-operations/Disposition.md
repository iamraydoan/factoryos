# Quality Operations — Disposition

> **ISA-95 Reference:** Part 3, §6 — Material Disposition Decision

## 1. Overview
**Disposition** is the formal decision made on a non-conforming unit or lot after an NCR is raised. It determines the physical fate of the material and the actions required before it can re-enter the production stream or be written off.

---

## 2. Disposition Types

| Disposition | Authority | Action Required |
|---|---|---|
| **Rework** | Quality Inspector | Unit routed to designated Rework Station; must pass re-inspection before acceptance |
| **Scrap** | Quality Inspector | Unit written off; removed from the active lot; scrap reason logged for OEE Quality metric |
| **Use-As-Is (UAI)** | Quality Manager (mandatory approval) | Unit accepted with documented deviation; requires engineering or customer sign-off |
| **Return to Supplier** | Quality Manager | Incoming material lot rejected and returned; supplier notified |

---

## 3. Rework Routing
When Rework is selected:
1. The non-conforming unit is assigned to the designated **Rework Work Center**.
2. A new Operation is created for the rework task (rework routing may differ from standard routing).
3. Upon rework completion, a fresh **Inspection** is automatically triggered.
4. If the re-inspection passes, the unit rejoins the production lot as a **reworked unit** (flagged in genealogy).
5. If it fails again, the NCR is re-opened for a new disposition decision.

---

## 4. Key Events

| Event | Description |
|---|---|
| `quality.disposition.rework_initiated` | Unit routed to Rework Work Center |
| `quality.disposition.scrapped` | Unit written off; OEE Quality counter decremented |
| `quality.disposition.use_as_is_approved` | Quality Manager approves UAI with documented deviation |
| `quality.disposition.return_to_supplier_initiated` | Lot flagged for supplier return |
