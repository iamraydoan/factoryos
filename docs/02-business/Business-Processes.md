# Business Domain: Core Manufacturing Business Processes

## 1. Overview
Describes the real-world operational workflows within a manufacturing facility that FactoryOS must support, from a pure business perspective (not technical implementation).

---

## 2. Core Business Processes

### 2.1 Production Execution
```text
[ ERP creates Production Order ]
        │
        ▼
[ Planning team schedules & assigns to Lines ]
        │
        ▼
[ Work Order dispatched to specific Line / Station ]
        │
        ▼
[ Operator scans QR → receives step-by-step work instructions ]
        │
        ▼
[ Executes each Operation following the Routing sequence ]
        │
        ▼
[ Work Order completed → Finished goods recorded ]
```

### 2.2 Quality Inspection
```text
[ QC Inspector receives inline inspection task at Station ]
        │
        ▼
[ Executes inspection checklist against quality standards ]
        │
   ┌────┴────┐
  PASS      FAIL
   │          │
   ▼          ▼
[ OK → ]  [ Raise NCR → Quarantine affected lot ]
               │
               ▼
    ┌──────────┴──────────┐
  REWORK               SCRAP
   │                     │
   ▼                     ▼
[ Re-process at     [ Write-off unit,
  designated          record scrap reason,
  rework station ]    update OEE Quality metric ]
               │
               ▼
       [ CAPA opened → Root cause analysis & corrective action ]
```

### 2.3 Maintenance
```text
[ Machine fault detected / OEE drops below threshold ]
        │
        ▼
[ Maintenance Work Order auto-created ]
        │
        ▼
[ Nearest available Technician dispatched ]
        │
        ▼
[ Technician acknowledges, repairs, closes Work Order ]
        │
        ▼
[ Asset Health & MTTR metrics updated ]
```

### 2.4 Material Flow & Traceability
```text
[ Raw material received at warehouse ]
        │
        ▼
[ Lot created & QR label printed ]
        │
        ▼
[ Staging: material allocated to Line / Station buffer ]
        │
        ▼
[ Operator scans QR to confirm correct material ]
        │
        ▼
[ Material consumed → Genealogy tree recorded ]
        │
        ▼
[ Finished product assigned As-Built Serial Number ]
```

---

## 3. Business Rules & Decisions

### 3.1 Unplanned / Emergency Work Order Approval
Emergency Work Orders (e.g., urgent customer re-run, field return rework) bypass the standard planning queue and follow an expedited approval path:
1. Production Supervisor raises an Emergency WO request in the system.
2. Plant Manager receives an approval notification and must confirm within **30 minutes**.
3. Upon approval, the WO is inserted at the top of the active queue for the target Line.
4. All bumped WOs are automatically re-sequenced and Supervisors are notified of the schedule change.

### 3.2 Scrap vs. Rework Decision after NCR
When an NCR is raised, the Quality Inspector selects one of three dispositions:
- **Rework:** The unit has a minor, correctable defect. It is routed to a designated Rework Station. Upon completion, it re-enters the inspection flow before being accepted.
- **Scrap:** The unit has a critical defect that cannot be corrected. It is written off, removed from the active lot, and the scrap reason is logged for OEE Quality degradation.
- **Use As-Is (UAI):** In exceptional cases, a unit that does not fully meet spec may be accepted with documented deviation and customer/engineering approval.

### 3.3 Maintenance Dispatch SLA Targets

| Fault Severity | SLA (Technician on-site) | Production Impact |
|---|---|---|
| **Critical** (full line stoppage) | ≤ 15 minutes | Line halted |
| **Major** (degraded performance) | ≤ 45 minutes | Reduced throughput |
| **Minor** (non-urgent observation) | ≤ 4 hours | No immediate impact |

### 3.4 Line Capacity Balancing
When multiple active Work Orders compete for the same Line capacity:
1. Work Orders are prioritized by **due date** first, then **customer priority tier** (A/B/C).
2. The Planning team uses the Finite Capacity Scheduling module to resolve conflicts.
3. Any WO that cannot be completed on time triggers an automatic **at-risk alert** to the Plant Manager.
4. Supervisors may manually override sequencing with documented justification.
