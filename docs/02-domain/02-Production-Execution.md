# Bounded Context: Production Order Execution & Dispatching

## 1. Overview
Manages work orders, operational routings, station execution tasks, shift schedules, and shop-floor operator dispatching.

---

## 2. Core Entities & Aggregates

* **WorkOrder:** Production target for a specific product quantity and standard routing.
* **Routing:** Step-by-step operational procedure required to manufacture a product.
* **Operation:** Individual manufacturing step assigned to a station or work center.
* **Shift:** Work schedule for human operators and equipment lines.

---

## 3. Work Order State Machine

```text
[ DRAFT ] ──> [ RELEASED ] ──> [ IN_PROGRESS ] ──> [ COMPLETED ]
                                      │
                                      └──> [ HELD / ON_PAUSE ]
```

---

## 4. Key Events

* `production.work_order.released`
* `production.operation.started`
* `production.operation.completed`
* `production.work_order.held`
