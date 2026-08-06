# Production Operations — Production Request

> **ISA-95 Reference:** Part 2, Section 4 — Operations Schedule / Production Request

## 1. Overview
The **Production Request** is the entry point from Level 4 (ERP) into Level 3 (FactoryOS). It represents the ERP system's demand signal: produce X quantity of Product Y, to be delivered by date Z. FactoryOS must accept, validate, and transform this request into a schedulable Production Schedule.

---

## 2. Entity Definitions

| Entity | ISA-95 Term | Description |
|---|---|---|
| **ProductionRequest** | Operations Request | Inbound demand from ERP: product, quantity, required-by date, priority |
| **RequestLine** | Operations Request Segment | Individual line item within a request for a specific MaterialDefinition |

---

## 3. Production Request State Machine

```text
[ RECEIVED ] ──▶ [ VALIDATED ] ──▶ [ SCHEDULED ] ──▶ [ FULFILLED ]
                      │
                      ▼
                 [ REJECTED ]  (missing BOM, material shortage, capacity unavailable)
```

---

## 4. Validation Rules
Before a request can be accepted:
- The `MaterialDefinition` referenced must exist and have an active BOM.
- Required materials must be available (or have a confirmed inbound date).
- Sufficient Work Center capacity must exist within the required delivery window.

---

## 5. Key Events

| Event | Description |
|---|---|
| `production.request.received` | ERP pushes a new Production Request into FactoryOS |
| `production.request.validated` | Request passes capacity and material checks |
| `production.request.rejected` | Request fails validation; ERP is notified with reason |
| `production.request.fulfilled` | All Work Orders for the request are completed |
