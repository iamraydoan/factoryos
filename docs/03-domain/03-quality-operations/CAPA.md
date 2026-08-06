# Quality Operations — CAPA

> **ISA-95 Reference:** Part 3, §6 — Corrective and Preventive Action

## 1. Overview
**CAPA (Corrective and Preventive Action)** is the structured problem-solving workflow triggered by recurring NCRs or systemic quality failures. It investigates root causes, implements corrective actions, and tracks their effectiveness to prevent recurrence.

---

## 2. Entity Definitions

| Entity | Description |
|---|---|
| **CAPA** | The formal investigation and action plan for a quality problem |
| **RootCauseAnalysis** | The 5-Why or Fishbone analysis record linking the symptom to the root cause |
| **CorrectiveAction** | A specific task assigned to an owner with a due date to fix the root cause |
| **PreventiveAction** | A system or process change to prevent recurrence in other areas |
| **EffectivenessReview** | A follow-up inspection after a defined period to confirm the CAPA is working |

---

## 3. CAPA State Machine

```text
[ OPENED ] ──▶ [ ROOT_CAUSE_IDENTIFIED ] ──▶ [ ACTIONS_ASSIGNED ]
                                                       │
                                                       ▼
                                             [ ACTIONS_IN_PROGRESS ]
                                                       │
                                                       ▼
                                           [ EFFECTIVENESS_REVIEW ]
                                              │              │
                                           Effective?    Not Effective?
                                              │              │
                                              ▼              ▼
                                         [ CLOSED ]   [ REOPENED ]
```

---

## 4. CAPA Triggers
A CAPA may be opened:
- Automatically, when an NCR of **Critical** severity is created.
- Manually by the Quality Manager for repeat defects (same defect class ≥ 3 times in 30 days).
- Following a customer complaint or field return.

---

## 5. Key Events

| Event | Description |
|---|---|
| `quality.capa.opened` | CAPA initiated from an NCR or manual trigger |
| `quality.capa.root_cause_identified` | Root Cause Analysis completed and recorded |
| `quality.capa.action_assigned` | A Corrective or Preventive Action task assigned to an owner |
| `quality.capa.effectiveness_confirmed` | Follow-up review confirms the CAPA resolved the issue |
| `quality.capa.closed` | All actions verified effective; CAPA formally closed |
