# ADR-0004: High-Throughput Telemetry Ingestion Architecture & Driver Strategy

* **Status:** Accepted
* **Date:** 2026-08-10
* **Authors:** Founding CTO, Principal Software Architect

---

## 1. Context

FactoryOS must ingest high-frequency time-series telemetry (up to 10,000+ signals/second) emitted from thousands of physical machines on the shop floor. 

Key technical requirements:
* Prevent database connection exhaustion under massive concurrent edge connections.
* Maximize TimescaleDB write throughput while minimizing database CPU utilization.
* Provide sub-second in-memory anomaly detection and stream processing without blocking transactional microservices.
* Maintain deterministic data retention (30 days raw data, 1 year continuous aggregates).

---

## 2. Decision

We adopt a **Decoupled Stream Ingestion & Micro-Batching Architecture** implemented in **Go** (`services/analytics-engine`):

1. **Language Ecosystem:** Implement `services/analytics-engine` and `platform/edge-runtime` in **Go (Golang)** to achieve low memory footprint (< 50MB RAM), sub-millisecond execution, and high concurrency.
2. **Edge-to-Cloud Transport Layer:** Edge IPC runtimes stream binary Protobuf `RecordBatch` payloads to the Cloud Ingestion Gateway via **gRPC over HTTP/2 (TLS on Port 443 / 50051)**.
3. **Internal Event Backbone:** The Cloud Ingestion Gateway authenticates edge nodes (mTLS/JWT) and produces validated messages to internal Apache Kafka topic `telemetry.raw.v1` with native **Snappy compression** inside the private cloud network.
4. **Micro-Batch Ingestion:** The ingestion consumer buffers incoming sensor readings into micro-batches (threshold: 500–1,000 records or 200ms periodic flush timer).
5. **Database Driver Strategy:** Use **`jackc/pgx/v5`** with **`pgx.CopyFrom` (PostgreSQL Binary COPY Protocol)** to stream batches directly into TimescaleDB hypertables (`raw_telemetry`), bypassing SQL query parsing and query planning overhead.
6. **Automated Lifecycle Policies:** Configure TimescaleDB 30-day raw data retention policies (`add_retention_policy`) and 365-day hourly continuous aggregates (`telemetry_hourly_summary`).

---

## 3. Consequences

### Positive Impacts
* **Network Isolation & Security:** Apache Kafka is strictly confined to internal private subnets, never exposed directly to public WAN/Internet.
* **Corporate Firewall Compatibility:** gRPC over HTTP/2 uses standard HTTPS (Port 443 / 50051), traversing enterprise factory proxies without requiring custom ports (e.g., 9092).
* **Extreme Write Throughput:** `pgx.CopyFrom` achieves 100,000–500,000+ records/second, 10x–50x faster than traditional SQL `INSERT` statements.
* **Database & Cluster Protection:** Microservices and Edge IPCs do not open direct database connections or raw Kafka sockets; connection pooling and rate-limiting are enforced at the Ingestion Gateway and `pgxpool`.
* **Real-Time Stream Alerts:** Threshold evaluation occurs in-memory before database persistence, enabling instantaneous alarm event emission.

### Negative Impacts & Trade-offs
* `pgx.CopyFrom` does not natively support `ON CONFLICT DO UPDATE` (Upsert). This trade-off is acceptable because sensor telemetry is strictly immutable, append-only time-series data.
* Introducing an Ingestion Gateway adds a lightweight proxy hop before Kafka, but this is necessary for edge authentication, firewall traversal, and internal cluster security.

---

## 4. Alternatives Considered

* **Direct Public Kafka Broker Exposure:** Rejected due to severe security risks (exposing internal event infrastructure to public WAN), complex multi-broker advertised listener management over dynamic factory IPs, and corporate firewall blocks on port 9092/9094.
* **HTTP/1.1 REST JSON Ingestion:** Rejected due to high bandwidth overhead (HTTP plain-text headers and JSON field repetition) and connection renegotiation overhead under high-frequency 10,000+ signals/sec.
* **Direct Edge-to-Database Ingestion:** Rejected due to severe security vulnerabilities (exposing database credentials on shop floor IPCs) and catastrophic connection exhaustion ("thundering herd" after network reconnection).
* **Heavyweight ORMs (GORM, Hibernate, Prisma):** Rejected due to massive object allocation, garbage collection overhead under high streaming throughput, and lack of binary COPY protocol optimization.
* **Multi-Row SQL INSERT Queries:** Rejected because the database engine must repeatedly parse, analyze, and plan large SQL text strings, causing high database CPU utilization.
