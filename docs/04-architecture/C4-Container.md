# Architecture Specification: C4 Container Diagram

## 1. Container Topology

```mermaid
graph TB
    subgraph Client Layer
        WebUI[Operator Terminal Frontend - React/NextJS]
    end

    subgraph API Layer
        Gateway[API & Ingestion Gateway - Traefik]
    end

    subgraph Core Microservices
        ProdSvc[Production Service - Java/Spring Boot]
        WhSvc[Warehouse Service - Java/Spring Boot]
        MaintSvc[Maintenance Service - Java/Spring Boot]
        QualSvc[Quality Service - Java/Spring Boot]
        AnalyticsSvc[Analytics Engine - Go]
    end

    subgraph Platform Infrastructure
        EventBus[Apache Kafka Event Bus (Private Subnet)]
        DB[(PostgreSQL + Outbox)]
        TSDB[(TimescaleDB Historian)]
        Cache[(Valkey - Redis Drop-in Replacement)]
        Observability[OpenTelemetry + Prometheus + Jaeger + Loki]
    end

    subgraph Edge Layer
        EdgeAgent[Factory Edge Agent - Go Edge Runtime]
    end

    WebUI -->|REST/GraphQL| Gateway
    Gateway --> ProdSvc
    Gateway --> WhSvc
    Gateway -->|Query OEE/Stats| AnalyticsSvc
    EdgeAgent -->|gRPC/HTTP2 over TLS| Gateway
    Gateway -->|Publish Telemetry Batch| EventBus
    EventBus -->|Consume Stream| AnalyticsSvc
    AnalyticsSvc -->|Batch CopyFrom| TSDB
    ProdSvc --> DB
    ProdSvc --> EventBus
```

---

## 2. Container Descriptions

* **Edge Agent:** Installed locally at factory sites. Ingests OPC-UA PLC signals and buffers data offline.
* **Production Service:** Manages station execution, state machine progression, and operator dispatch.
* **Warehouse Service:** Tracks material lots and genealogy as-built component assembly.
* **Kafka Event Bus:** Asynchronous backbone routing domain events across microservices.
