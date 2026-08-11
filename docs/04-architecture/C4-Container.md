# Architecture Specification: C4 Container Diagram

## 1. Container Topology

```mermaid
graph TB
    subgraph Client Layer
        WebUI["Operator Terminal Frontend - React/NextJS"]
    end

    subgraph API Layer
        Gateway["API & Ingestion Gateway - Traefik"]
    end

    subgraph Core Microservices
        IngestionSvc["Telemetry Ingestion Service - Go"]
        AnalyticsSvc["Analytics Engine - Go"]
        ProdSvc["Production Service - Java/Spring Boot"]
        WhSvc["Warehouse Service - Java/Spring Boot"]
        MaintSvc["Maintenance Service - Java/Spring Boot"]
        QualSvc["Quality Service - Java/Spring Boot"]
    end

    subgraph Platform Infrastructure
        EventBus["Apache Kafka Event Bus (Private Subnet)"]
        DB[("PostgreSQL + Outbox")]
        TSDB[("TimescaleDB Historian")]
        Cache[("Valkey - Redis Drop-in Replacement")]
        Observability["OpenTelemetry + Prometheus + Jaeger + Loki"]
    end

    subgraph Edge Layer
        EdgeAgent["Factory Edge Agent - Go Edge Runtime"]
    end

    WebUI -->|REST/GraphQL| Gateway
    Gateway --> ProdSvc
    Gateway --> WhSvc
    Gateway -->|Query OEE/Stats| AnalyticsSvc
    EdgeAgent -->|gRPC Stream over TLS| Gateway
    Gateway -->|Forward gRPC| IngestionSvc
    IngestionSvc -->|Produce Snappy Batch| EventBus
    EventBus -->|Consume Stream| AnalyticsSvc
    AnalyticsSvc -->|Batch CopyFrom| TSDB
    ProdSvc --> DB
    ProdSvc --> EventBus
```

---

## 2. Container Descriptions

* **Edge Agent:** Installed locally at factory sites. Ingests OPC-UA PLC signals and buffers data in local SQLite when offline.
* **API Gateway (Traefik):** Single public entrypoint for TLS termination, JWT verification, and routing REST/gRPC traffic.
* **Telemetry Ingestion Service:** High-throughput Go gRPC service that authenticates edge nodes, validates Protobuf batches, and produces them into Kafka `telemetry.raw.v1`.
* **Analytics Engine:** Consumes real-time streams from Kafka, evaluates dynamic thresholds in memory, calculates OEE, and streams micro-batches into TimescaleDB via `pgx.CopyFrom`.
* **Production Service:** Manages station execution, state machine progression, and operator dispatch.
* **Warehouse Service:** Tracks material lots and genealogy as-built component assembly.
* **Kafka Event Bus:** Asynchronous backbone routing domain events across microservices in private cloud subnets.
