# Production Operations — Production Schedule

> **ISA-95 Reference:** Part 2, Section 4 — Operations Schedule

## 1. Overview
The **Production Schedule** is the Planning team's answer to a Production Request. It assigns Work Orders to specific Work Centers across a defined time horizon, resolving capacity conflicts and sequencing priorities.

---

## 2. Entity Definitions

| Entity | ISA-95 Term | Description |
|---|---|---|
| **ProductionSchedule** | Operations Schedule | The master time-bound plan for a set of Work Orders across Work Centers |
| **ScheduleEntry** | Operations Request Segment | Assignment of one Work Order to a specific Work Center within a time window |

---

## 3. Scheduling Logic

### Priority Rules (in order)
1. Customer priority tier (A > B > C)
2. Required-by date (earliest first)
3. Manual Supervisor override (with documented justification)

### Capacity Conflict Resolution
- If two Work Orders compete for the same Work Center in the same window, the lower-priority WO is pushed to the next available slot.
- If no slot exists within the required delivery window, an **at-risk alert** is raised to the Plant Manager.

---

## 4. Schedule State Machine

```text
[ DRAFT ] ──▶ [ PUBLISHED ] ──▶ [ IN_EXECUTION ] ──▶ [ CLOSED ]
                   │
                   ▼
             [ REVISED ]  (re-planning triggered by Emergency WO or capacity change)
```

---

## 5. Key Events

| Event | Description |
|---|---|
| `production.schedule.published` | A schedule is approved and dispatched to the floor |
| `production.schedule.revised` | Schedule is re-planned due to priority change or capacity event |
| `production.schedule.work_order_at_risk` | A WO cannot meet its due date; Plant Manager alerted |
