# Architecture Specification: C4 Container Diagram

## 1. Container Topology

```mermaid
graph TB
    subgraph Client Layer
        WebUI[Operator Terminal Frontend - React/NextJS]
    end

    subgraph API Layer
        Gateway[API Gateway - Envoy / Traefik]
    end

    subgraph Core Microservices
        ProdSvc[Production Service]
        WhSvc[Warehouse Service]
        MaintSvc[Maintenance Service]
        QualSvc[Quality Service]
    end

    subgraph Platform Infrastructure
        EventBus[Kafka / Redpanda Event Bus]
        DB[(PostgreSQL + Outbox)]
        TSDB[(TimescaleDB Historian)]
    end

    subgraph Edge Layer
        EdgeAgent[Factory Edge Agent - Go Edge Runtime]
    end

    WebUI --> Gateway
    Gateway --> ProdSvc
    Gateway --> WhSvc
    EdgeAgent -->|MQTT/Protobuf| EventBus
    ProdSvc --> DB
    ProdSvc --> EventBus
    EventBus --> TSDB
```

---

## 2. Container Descriptions

* **Edge Agent:** Installed locally at factory sites. Ingests OPC-UA PLC signals and buffers data offline.
* **Production Service:** Manages station execution, state machine progression, and operator dispatch.
* **Warehouse Service:** Tracks material lots and genealogy as-built component assembly.
* **Kafka Event Bus:** Asynchronous backbone routing domain events across microservices.
