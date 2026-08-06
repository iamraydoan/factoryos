# Architecture Specification: C4 Container Diagram

## 1. Container Topology

```mermaid
graph TB
    subgraph Client Layer
        WebUI[Operator Terminal Frontend - React/NextJS]
    end

    subgraph API Layer
        Gateway[API Gateway - Traefik]
    end

    subgraph Core Microservices
        ProdSvc[Production Service - Java/Spring Boot]
        WhSvc[Warehouse Service - Java/Spring Boot]
        MaintSvc[Maintenance Service - Java/Spring Boot]
        QualSvc[Quality Service - Java/Spring Boot]
    end

    subgraph Platform Infrastructure
        EventBus[Apache Kafka Event Bus]
        DB[(PostgreSQL + Outbox)]
        TSDB[(TimescaleDB Historian)]
        Cache[(Valkey - Redis Drop-in Replacement)]
        Observability[OpenTelemetry + Prometheus + Jaeger + Loki]
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
