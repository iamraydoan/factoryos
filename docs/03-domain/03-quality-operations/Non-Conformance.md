# Quality Operations — Non-Conformance (NCR)

> **ISA-95 Reference:** Part 3, §6 — Quality Non-Conformance

## 1. Overview
A **Non-Conformance Report (NCR)** is raised when an Inspection fails or a defect is detected during production. It captures the defect, triggers automatic lot quarantine, and initiates the Disposition decision.

---

## 2. Entity Definitions

| Entity | Description |
|---|---|
| **NonConformanceReport (NCR)** | The formal record of a quality failure, linked to a specific Inspection, Lot, and Operation |
| **Defect** | The specific deviation from specification detected (one NCR may contain multiple Defects) |
| **DefectClass** | Classification of defect type (e.g., Dimensional, Surface, Functional, Assembly) |
| **QuarantineRecord** | Automatic hold placed on the affected Lot upon NCR creation |

---

## 3. NCR State Machine

```text
[ OPEN ] ──▶ [ UNDER_REVIEW ] ──▶ [ DISPOSITION_SELECTED ]
                                           │
                         ┌─────────────────┼──────────────────┐
                         ▼                 ▼                   ▼
                    [ REWORK ]       [ SCRAPPED ]        [ USE_AS_IS ]
                         │                                     │
                         ▼                                     ▼
                   [ RE-INSPECTED ]                   [ DEVIATION_APPROVED ]
                         │
                    Pass? │ Fail?
                          │──▶ [ RE-OPENED ]
                          │
                          ▼
                    [ CLOSED ]
```

---

## 4. Defect Classification

| Class | Examples |
|---|---|
| **Dimensional** | Out-of-tolerance hole diameter, incorrect length |
| **Surface** | Scratch, dent, corrosion, incomplete coating |
| **Functional** | Electrical failure, leak test failure |
| **Assembly** | Missing component, wrong component, incorrect orientation |
| **Documentation** | Label missing, traveler incomplete |

---

## 5. Key Events

| Event | Description |
|---|---|
| `quality.ncr.raised` | NCR created from a failed Inspection; quarantine applied to affected Lot |
| `quality.ncr.disposition_selected` | Quality Inspector or Manager selects Rework / Scrap / UAI |
| `quality.ncr.closed` | NCR resolved and closed after passing re-inspection or approved disposition |
