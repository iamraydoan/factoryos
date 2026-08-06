# System Architecture Overview

## 1. High-Level Architecture Topology

```
+-------------------------------------------------------------------------------+
|                           Enterprise Applications (ERP / CRM)                  |
+-------------------------------------------------------------------------------+
                                       ▲
                                       │ gRPC / OpenAPI REST Gateway
                                       ▼
+-------------------------------------------------------------------------------+
|                       FactoryOS Cloud Control Plane                           |
|                                                                               |
|  +---------------------+  +---------------------+  +-----------------------+  |
|  | Production Service  |  | Warehouse Service   |  | Quality Service       |  |
|  +---------------------+  +---------------------+  +-----------------------+  |
|  +---------------------+  +---------------------+  +-----------------------+  |
|  | Maintenance Service |  | Planning Service    |  | Analytics Engine      |  |
|  +---------------------+  +---------------------+  +-----------------------+  |
|                                                                               |
|  +-------------------------------------------------------------------------+  |
|  | Event Platform (Apache Kafka & Schema Registry)                                 |  |
|  +-------------------------------------------------------------------------+  |
|  | Durable Workflow Engine (Temporal.io)                                           |  |
|  +-------------------------------------------------------------------------+  |
|  | Identity & Access Management (Zitadel - OIDC/RBAC)                      |  |
|  +-------------------------------------------------------------------------+  |
+-------------------------------------------------------------------------------+
                                       ▲
                                       │ TLS 1.3 / gRPC Tunnel / MQTT Bridge
                                       ▼
+-------------------------------------------------------------------------------+
|                         Factory Edge Nodes (Local Network)                    |
|  +-------------------------------------------------------------------------+  |
|  | FactoryOS Edge Runtime (Offline Store & Forward Engine, Protocol Drivers) |  |
|  +-------------------------------------------------------------------------+  |
|         ▲                                   ▲                    ▲            |
|         │ OPC-UA                            │ MQTT               │ Modbus     |
|  +--------------+                    +--------------+     +--------------+    |
|  | CNC Machine  |                    | Robotic Arm  |     | PLC Sensors  |    |
|  +--------------+                    +--------------+     +--------------+    |
+-------------------------------------------------------------------------------+
```

---

## 2. Key Architecture Principles

1. **Schema-First API Contracts:** All inter-service communications use versioned gRPC Protobuf or AsyncAPI schemas defined centrally in `api-contracts`.
2. **Offline-First Edge Autonomy:** Factory edge nodes must operate autonomously during cloud connectivity loss for at least 72 hours.
3. **Decoupled Business Services:** Domain microservices interact solely via domain event publishing and durable workflow orchestrations.
4. **Immutable Audit Ledger:** All physical actions, operator sign-offs, and state transitions are stored in append-only immutable event logs.
